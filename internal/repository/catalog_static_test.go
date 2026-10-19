package repository

import "testing"

func TestStaticCatalog(t *testing.T) {
	catalog := DefaultStaticCatalog("", "sample-producer")
	if category, ok := catalog.ProductCategory("transfer_bifast"); !ok || category != "transfer" {
		t.Fatalf("ProductCategory() = %q, %v", category, ok)
	}
	if catalog.ValidProvider("pulsa", "telkomsel") != true {
		t.Fatal("expected telkomsel to be valid")
	}
	if catalog.ProviderName("bca") != "BCA" {
		t.Fatalf("unexpected provider name: %q", catalog.ProviderName("bca"))
	}
	if !catalog.ValidPartner("sample_partner") || catalog.ValidPartner("other") {
		t.Fatal("unexpected partner validation result")
	}
	if !catalog.AllowedSource("sample-producer") || catalog.AllowedSource("unknown") {
		t.Fatal("unexpected source validation result")
	}
	if !catalog.ValidApp("app_a") || catalog.ValidApp("unknown") {
		t.Fatal("unexpected app validation result")
	}
}
