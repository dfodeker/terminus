package gid

import (
	"strings"
	"testing"
	"time"
)

func TestGenerator(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("NewGenerator failed: %v", err)
	}

	// Generate multiple IDs and ensure uniqueness
	ids := make(map[uint64]bool)
	for i := 0; i < 10000; i++ {
		id := gen.Generate()
		if ids[id] {
			t.Errorf("duplicate ID generated: %d", id)
		}
		ids[id] = true
	}
}

func TestGeneratorInvalidMachineID(t *testing.T) {
	_, err := NewGenerator(1024)
	if err == nil {
		t.Error("expected error for machine ID > 1023")
	}
}

func TestExtractComponents(t *testing.T) {
	gen, _ := NewGenerator(42)
	before := time.Now().Truncate(time.Millisecond)
	id := gen.Generate()
	after := time.Now().Add(time.Millisecond).Truncate(time.Millisecond)

	extractedTime := ExtractTime(id)
	if extractedTime.Before(before) || extractedTime.After(after) {
		t.Errorf("extracted time %v not between %v and %v", extractedTime, before, after)
	}

	machineID := ExtractMachineID(id)
	if machineID != 42 {
		t.Errorf("expected machine ID 42, got %d", machineID)
	}
}

func TestGIDString(t *testing.T) {
	g := ProductGID(123456789)
	expected := "gid://mystoreos/Product/123456789"
	if g.String() != expected {
		t.Errorf("expected %q, got %q", expected, g.String())
	}
}

func TestGIDParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    GID
		wantErr bool
	}{
		{
			name:  "valid product GID",
			input: "gid://mystoreos/Product/123456789",
			want:  GID{Type: EntityProduct, ID: 123456789},
		},
		{
			name:  "valid store GID",
			input: "gid://mystoreos/Store/999",
			want:  GID{Type: EntityStore, ID: 999},
		},
		{
			name:  "valid tenant GID",
			input: "gid://mystoreos/Tenant/1",
			want:  GID{Type: EntityTenant, ID: 1},
		},
		{
			name:    "missing prefix",
			input:   "Product/123",
			wantErr: true,
		},
		{
			name:    "wrong namespace",
			input:   "gid://shopify/Product/123",
			wantErr: true,
		},
		{
			name:    "invalid entity type",
			input:   "gid://mystoreos/Invalid/123",
			wantErr: true,
		},
		{
			name:    "invalid ID",
			input:   "gid://mystoreos/Product/abc",
			wantErr: true,
		},
		{
			name:    "missing ID",
			input:   "gid://mystoreos/Product/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGIDBase64Roundtrip(t *testing.T) {
	original := ProductVariantGID(9876543210)
	encoded := original.Base64()

	decoded, err := ParseBase64(encoded)
	if err != nil {
		t.Fatalf("ParseBase64 failed: %v", err)
	}

	if decoded != original {
		t.Errorf("roundtrip failed: got %v, want %v", decoded, original)
	}
}

func TestGIDHelpers(t *testing.T) {
	tests := []struct {
		gid      GID
		wantType EntityType
	}{
		{ProductGID(1), EntityProduct},
		{ProductVariantGID(2), EntityProductVariant},
		{StoreGID(3), EntityStore},
		{TenantGID(4), EntityTenant},
		{UserGID(5), EntityUser},
		{RoleGID(6), EntityRole},
		{PermissionGID(7), EntityPermission},
		{CustomDomainGID(8), EntityCustomDomain},
	}

	for _, tt := range tests {
		if tt.gid.Type != tt.wantType {
			t.Errorf("helper created wrong type: got %s, want %s", tt.gid.Type, tt.wantType)
		}
	}
}

func TestEntityTypeIsValid(t *testing.T) {
	validTypes := []EntityType{
		EntityProduct, EntityProductVariant, EntityStore,
		EntityTenant, EntityUser, EntityRole, EntityPermission,
		EntityCustomDomain,
	}

	for _, et := range validTypes {
		if !et.IsValid() {
			t.Errorf("%s should be valid", et)
		}
	}

	invalidType := EntityType("Invalid")
	if invalidType.IsValid() {
		t.Error("Invalid should not be valid")
	}
}

func TestGIDIsZero(t *testing.T) {
	var zero GID
	if !zero.IsZero() {
		t.Error("zero GID should return IsZero() = true")
	}

	nonZero := ProductGID(1)
	if nonZero.IsZero() {
		t.Error("non-zero GID should return IsZero() = false")
	}
}

func TestMustParsePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse should panic on invalid input")
		}
	}()

	MustParse("invalid")
}

func TestMustParseSuccess(t *testing.T) {
	g := MustParse("gid://mystoreos/Product/123")
	if g.Type != EntityProduct || g.ID != 123 {
		t.Errorf("unexpected GID: %v", g)
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen, _ := NewGenerator(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate()
	}
}

func BenchmarkParse(b *testing.B) {
	input := "gid://mystoreos/Product/123456789"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(input)
	}
}

func TestConcurrentGeneration(t *testing.T) {
	gen, _ := NewGenerator(1)
	ids := make(chan uint64, 10000)

	// Generate IDs concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				ids <- gen.Generate()
			}
		}()
	}

	// Collect and check for duplicates
	seen := make(map[uint64]bool)
	for i := 0; i < 10000; i++ {
		id := <-ids
		if seen[id] {
			t.Errorf("duplicate ID: %d", id)
		}
		seen[id] = true
	}
}

func TestParseAllEntityTypes(t *testing.T) {
	entityTypes := []EntityType{
		EntityProduct, EntityProductVariant, EntityStore,
		EntityTenant, EntityUser, EntityRole, EntityPermission,
		EntityCustomDomain,
	}

	for _, et := range entityTypes {
		gidStr := "gid://mystoreos/" + string(et) + "/12345"
		g, err := Parse(gidStr)
		if err != nil {
			t.Errorf("failed to parse %s: %v", gidStr, err)
			continue
		}
		if g.Type != et {
			t.Errorf("parsed type %s, expected %s", g.Type, et)
		}
		if g.ID != 12345 {
			t.Errorf("parsed ID %d, expected 12345", g.ID)
		}
	}
}

func TestGIDStringContainsNoSpaces(t *testing.T) {
	g := ProductGID(123)
	if strings.Contains(g.String(), " ") {
		t.Error("GID string should not contain spaces")
	}
}
