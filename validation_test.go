package libmozhi

import "testing"

func TestValidateLanguagePairUsesTargetList(t *testing.T) {
	err := validateLanguagePair(langListDeepl("sl"), langListDeepl("tl"), "en", "pt-br")
	if err != nil {
		t.Fatalf("expected pt-br to be valid for DeepL targets: %v", err)
	}
}

func TestValidateLanguagePairRejectsTargetOnlySource(t *testing.T) {
	err := validateLanguagePair(langListDeepl("sl"), langListDeepl("tl"), "pt-br", "en")
	if err == nil || err.Error() != "Source language code invalid" {
		t.Fatalf("expected source validation error, got %v", err)
	}
}
