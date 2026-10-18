package appstore

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/config"

	"github.com/golang-jwt/jwt/v5"
)

// verifyAndDecodeJWS verifies an Apple JWS (compact serialization) and unmarshals
// its payload into out.
//
// Apple signs with ES256 and places the signing certificate chain in the JWS
// header's x5c field (leaf, intermediate, root — base64 DER). Verification:
//  1. parse the x5c leaf certificate and use its ECDSA public key to verify the
//     ES256 signature over header.payload;
//  2. verify the certificate chain. When APPLE_ROOT_CA_PEM is configured the
//     chain is verified up to that pinned root; otherwise the chain's internal
//     consistency (leaf←intermediate←root) is checked and a warning is logged
//     that the trust root is not pinned (hardening tracked in DAL-164).
func verifyAndDecodeJWS(token string, out interface{}) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ErrInvalidSignature
	}

	leaf, err := certChainFromHeader(parts[0])
	if err != nil {
		return err
	}
	pub, ok := leaf.chain[0].PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return ErrInvalidSignature
	}

	// 1. Signature over the ES256 token using the leaf's public key.
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"ES256"}))
	if _, err := parser.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return pub, nil
	}); err != nil {
		return ErrInvalidSignature
	}

	// 2. Certificate chain.
	if err := leaf.verifyChain(); err != nil {
		return err
	}

	// Payload → out.
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ErrInvalidSignature
	}
	return json.Unmarshal(payload, out)
}

type parsedChain struct {
	chain []*x509.Certificate // [leaf, intermediate, root]
}

func certChainFromHeader(encodedHeader string) (*parsedChain, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encodedHeader)
	if err != nil {
		return nil, ErrInvalidSignature
	}
	var hdr struct {
		Alg string   `json:"alg"`
		X5c []string `json:"x5c"`
	}
	if err := json.Unmarshal(raw, &hdr); err != nil {
		return nil, ErrInvalidSignature
	}
	if hdr.Alg != "ES256" || len(hdr.X5c) == 0 {
		return nil, ErrInvalidSignature
	}
	var chain []*x509.Certificate
	for _, b64 := range hdr.X5c {
		der, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, ErrInvalidSignature
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, ErrInvalidSignature
		}
		chain = append(chain, cert)
	}
	return &parsedChain{chain: chain}, nil
}

func (p *parsedChain) verifyChain() error {
	if len(p.chain) == 0 {
		return ErrInvalidSignature
	}
	leaf := p.chain[0]
	intermediates := x509.NewCertPool()
	for _, c := range p.chain[1:] {
		intermediates.AddCert(c)
	}

	roots := x509.NewCertPool()
	pinned := false
	if pem := config.GetConfig().AppleRootCAPEM; strings.TrimSpace(pem) != "" {
		if roots.AppendCertsFromPEM([]byte(pem)) {
			pinned = true
		}
	}
	if !pinned {
		// No pinned root: fall back to trusting the chain's own top cert so the
		// internal leaf←intermediate←root links are still verified. This checks
		// structural validity but NOT that the root is really Apple's — must be
		// pinned in staging/production.
		logger.WarningNoCtx("apple JWS: APPLE_ROOT_CA_PEM not set — verifying chain without a pinned trust root")
		roots.AddCert(p.chain[len(p.chain)-1])
	}

	if _, err := leaf.Verify(x509.VerifyOptions{
		Intermediates: intermediates,
		Roots:         roots,
		CurrentTime:   time.Now(),
	}); err != nil {
		return ErrInvalidSignature
	}
	return nil
}
