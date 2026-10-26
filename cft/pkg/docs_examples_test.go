package pkg_test

import "testing"

// The documentation site includes these templates verbatim, so a failing
// test here means a published example is wrong.

func TestDocsEmbed(t *testing.T) {
	runTest("docs/embed", t)
}

func TestDocsInclude(t *testing.T) {
	runTest("docs/include", t)
}

func TestDocsEnv(t *testing.T) {
	t.Setenv("DEPLOYED_BY", "ci-pipeline")
	runTest("docs/env", t)
}

func TestDocsConstants(t *testing.T) {
	runTest("docs/constants", t)
}

func TestDocsModule(t *testing.T) {
	runTest("docs/module", t)
}

func TestDocsForEach(t *testing.T) {
	runTest("docs/foreach", t)
}

func TestDocsIfParam(t *testing.T) {
	runTest("docs/ifparam", t)
}
