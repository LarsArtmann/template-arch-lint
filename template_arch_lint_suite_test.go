package template_arch_lint_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTemplateArchLint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TemplateArchLint Suite")
}
