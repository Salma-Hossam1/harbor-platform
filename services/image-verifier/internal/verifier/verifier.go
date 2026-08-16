package verifier

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var digestReferencePattern = regexp.MustCompile(
	`^[^@\s]+@sha256:[a-f0-9]{64}$`,
)

type Verifier struct {
	cosignBinary string
	publicKey    string
	registry     string
	username     string
	password     string
	ignoreTLog   bool
}

func New(
	cosignBinary string,
	publicKey string,
	registry string,
	username string,
	password string,
	ignoreTLog bool,
) *Verifier {
	return &Verifier{
		cosignBinary: cosignBinary,
		publicKey:    publicKey,
		registry:     registry,
		username:     username,
		password:     password,
		ignoreTLog:   ignoreTLog,
	}
}

func (v *Verifier) Verify(ctx context.Context, image string) error {
	if !digestReferencePattern.MatchString(image) {
		return fmt.Errorf("image must use an exact sha256 digest")
	}
	if !strings.HasPrefix(image, v.registry+"/") {
		return fmt.Errorf("image does not belong to configured Harbor registry")
	}

	args := []string{
		"verify",
		"--key", v.publicKey,
		"--registry-cacert", "/etc/harbor-ca/ca.crt",
		"--registry-username", v.username,
		"--registry-password", v.password,
	}

	if v.ignoreTLog {
		args = append(args, "--insecure-ignore-tlog")
	}

	args = append(args, image)

	cmd := exec.CommandContext(ctx, v.cosignBinary, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"cosign verification failed: %s",
			strings.TrimSpace(string(output)),
		)
	}

	return nil
}