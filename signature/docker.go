// Note: Consider the API unstable until the code supports at least three different image formats or transports.

package signature

import (
	"fmt"

	"github.com/projectatomic/skopeo/docker/utils"
)

// SignDockerManifest returns a signature for manifest as the specified dockerReference,
// using mech and keyIdentity.
func SignDockerManifest(manifest []byte, dockerReference string, mech SigningMechanism, keyIdentity string) ([]byte, error) {
	manifestDigest, err := utils.ManifestDigest(manifest)
	if err != nil {
		return nil, err
	}
	sig := privateSignature{
		Signature{
			DockerManifestDigest: manifestDigest,
			DockerReference:      dockerReference,
		},
	}
	return sig.sign(mech, keyIdentity)
}

// VerifyDockerManifestSignature checks that unverifiedSignature uses expectedKeyIdentity to sign unverifiedManifest as expectedDockerReference,
// using mech.
func VerifyDockerManifestSignature(unverifiedSignature, unverifiedManifest []byte,
	expectedDockerReference string, mech SigningMechanism, expectedKeyIdentity string) (*Signature, error) {
	sig, err := verifyAndExtractSignature(mech, unverifiedSignature, signatureAcceptanceRules{
		validateKeyIdentity: func(keyIdentity string) error {
			if keyIdentity != expectedKeyIdentity {
				return InvalidSignatureError{msg: fmt.Sprintf("Signature by %s does not match expected fingerprint %s", keyIdentity, expectedKeyIdentity)}
			}
			return nil
		},
		validateSignedDockerReference: func(signedDockerReference string) error {
			if signedDockerReference != expectedDockerReference {
				return InvalidSignatureError{msg: fmt.Sprintf("Docker reference %s does not match %s",
					signedDockerReference, expectedDockerReference)}
			}
			return nil
		},
		validateSignedDockerManifestDigest: func(signedDockerManifestDigest string) error {
			matches, err := utils.ManifestMatchesDigest(unverifiedManifest, signedDockerManifestDigest)
			if err != nil {
				return err
			}
			if !matches {
				return InvalidSignatureError{msg: fmt.Sprintf("Signature for docker digest %s does not match", signedDockerManifestDigest, signedDockerManifestDigest)}
			}
			return nil
		},
	})
	if err != nil {
		return nil, err
	}
	return sig, nil
}
