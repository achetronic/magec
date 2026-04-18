// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package msgutil

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"testing"

	"google.golang.org/adk/artifact"
	"google.golang.org/genai"
)

// fakeArtifactService is a minimal artifact.Service implementation for tests.
type fakeArtifactService struct {
	mu       sync.Mutex
	saved    []artifact.SaveRequest
	saveErr  error
	versions map[string]int64
}

func (f *fakeArtifactService) Save(_ context.Context, req *artifact.SaveRequest) (*artifact.SaveResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	if f.versions == nil {
		f.versions = map[string]int64{}
	}
	f.versions[req.FileName]++
	f.saved = append(f.saved, *req)
	return &artifact.SaveResponse{Version: f.versions[req.FileName]}, nil
}

func (f *fakeArtifactService) Load(_ context.Context, _ *artifact.LoadRequest) (*artifact.LoadResponse, error) {
	return &artifact.LoadResponse{Part: &genai.Part{Text: ""}}, nil
}
func (f *fakeArtifactService) Delete(_ context.Context, _ *artifact.DeleteRequest) error {
	return nil
}
func (f *fakeArtifactService) List(_ context.Context, _ *artifact.ListRequest) (*artifact.ListResponse, error) {
	return &artifact.ListResponse{}, nil
}
func (f *fakeArtifactService) Versions(_ context.Context, _ *artifact.VersionsRequest) (*artifact.VersionsResponse, error) {
	return &artifact.VersionsResponse{}, nil
}

// ----- ShouldInline -----

func TestShouldInline_BelowThreshold(t *testing.T) {
	if !ShouldInline(500 * 1024) {
		t.Fatalf("500 KiB should be inline")
	}
}

func TestShouldInline_AtThreshold(t *testing.T) {
	if !ShouldInline(DefaultInlineThreshold) {
		t.Fatalf("exactly %d bytes (threshold) should be inline", DefaultInlineThreshold)
	}
}

func TestShouldInline_AboveThreshold(t *testing.T) {
	if ShouldInline(DefaultInlineThreshold + 1) {
		t.Fatalf("threshold+1 bytes should NOT be inline")
	}
}

func TestShouldInline_Zero(t *testing.T) {
	if !ShouldInline(0) {
		t.Fatalf("0-byte file should be inline (degenerate but harmless)")
	}
}

// ----- InlinePart -----

func TestInlinePart_EncodesMimeAndBase64(t *testing.T) {
	part := InlinePart("image/png", []byte("hello"))

	inline, ok := part["inlineData"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected inlineData key to be a map, got %T", part["inlineData"])
	}
	if inline["mimeType"] != "image/png" {
		t.Fatalf("wrong mimeType: %v", inline["mimeType"])
	}
	want := base64.StdEncoding.EncodeToString([]byte("hello"))
	if inline["data"] != want {
		t.Fatalf("wrong data: got %v want %v", inline["data"], want)
	}
}

func TestInlinePart_EmptyMimeFallsBackToOctetStream(t *testing.T) {
	part := InlinePart("", []byte("x"))
	inline := part["inlineData"].(map[string]interface{})
	if inline["mimeType"] != "application/octet-stream" {
		t.Fatalf("empty mime should fall back to octet-stream, got %v", inline["mimeType"])
	}
}

// ----- StoreAsArtifact -----

func TestStoreAsArtifact_SavesAndReturnsDescriptor(t *testing.T) {
	svc := &fakeArtifactService{}
	line, err := StoreAsArtifact(context.Background(), svc, "app", "user", "sess", "big.pdf", "application/pdf", make([]byte, 2*1024*1024))
	if err != nil {
		t.Fatalf("StoreAsArtifact: %v", err)
	}
	if len(svc.saved) != 1 {
		t.Fatalf("expected 1 artifact saved, got %d", len(svc.saved))
	}
	req := svc.saved[0]
	if req.AppName != "app" || req.UserID != "user" || req.SessionID != "sess" || req.FileName != "big.pdf" {
		t.Fatalf("save request fields wrong: %+v", req)
	}
	if req.Part == nil || req.Part.InlineData == nil || req.Part.InlineData.MIMEType != "application/pdf" {
		t.Fatalf("part not built correctly: %+v", req.Part)
	}
	if !strings.Contains(line, "big.pdf") || !strings.Contains(line, "application/pdf") || !strings.Contains(line, "MB") {
		t.Fatalf("descriptor line missing expected parts: %q", line)
	}
}

func TestStoreAsArtifact_NilServiceErrors(t *testing.T) {
	_, err := StoreAsArtifact(context.Background(), nil, "app", "user", "sess", "f.bin", "application/octet-stream", []byte("x"))
	if err == nil {
		t.Fatalf("expected error when service is nil")
	}
}

func TestStoreAsArtifact_EmptyFilenameErrors(t *testing.T) {
	svc := &fakeArtifactService{}
	_, err := StoreAsArtifact(context.Background(), svc, "app", "user", "sess", "", "application/octet-stream", []byte("x"))
	if err == nil {
		t.Fatalf("expected error when filename is empty")
	}
	if len(svc.saved) != 0 {
		t.Fatalf("nothing should have been saved on validation error")
	}
}

func TestStoreAsArtifact_PropagatesSaveError(t *testing.T) {
	svc := &fakeArtifactService{saveErr: errors.New("disk full")}
	_, err := StoreAsArtifact(context.Background(), svc, "app", "user", "sess", "f.bin", "application/octet-stream", []byte("x"))
	if err == nil {
		t.Fatalf("expected error to propagate")
	}
	if !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("underlying error message lost: %v", err)
	}
}

func TestStoreAsArtifact_EmptyMimeFallsBack(t *testing.T) {
	svc := &fakeArtifactService{}
	line, err := StoreAsArtifact(context.Background(), svc, "app", "user", "sess", "x.bin", "", []byte("data"))
	if err != nil {
		t.Fatalf("StoreAsArtifact: %v", err)
	}
	if !strings.Contains(line, "application/octet-stream") {
		t.Fatalf("expected fallback mime in descriptor, got %q", line)
	}
	if svc.saved[0].Part.InlineData.MIMEType != "application/octet-stream" {
		t.Fatalf("fallback mime not applied on Save")
	}
}

// ----- AttachedArtifactsBlock -----

func TestAttachedArtifactsBlock_Empty(t *testing.T) {
	if AttachedArtifactsBlock(nil) != "" {
		t.Fatalf("nil input must produce empty block")
	}
	if AttachedArtifactsBlock([]string{}) != "" {
		t.Fatalf("empty slice must produce empty block")
	}
}

func TestAttachedArtifactsBlock_WrapsLines(t *testing.T) {
	block := AttachedArtifactsBlock([]string{"- foo.pdf (application/pdf, 3.2 MB)", "- bar.png (image/png, 1.1 MB)"})
	if !strings.HasPrefix(strings.TrimLeft(block, "\n"), "<!--MAGEC_ATTACHED_ARTIFACTS:") {
		t.Fatalf("block must start with the MAGEC_ATTACHED_ARTIFACTS marker: %q", block)
	}
	if !strings.HasSuffix(block, ":MAGEC_ATTACHED_ARTIFACTS-->") {
		t.Fatalf("block must end with the closing marker: %q", block)
	}
	if !strings.Contains(block, "foo.pdf") || !strings.Contains(block, "bar.png") {
		t.Fatalf("descriptor lines missing: %q", block)
	}
	if !strings.Contains(block, "load_artifact") {
		t.Fatalf("block must instruct the model to call load_artifact: %q", block)
	}
}

// ----- humanSize (private, covered indirectly but one direct test for edges) -----

func TestHumanSize_Boundaries(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tc := range cases {
		if got := humanSize(tc.in); got != tc.want {
			t.Fatalf("humanSize(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
