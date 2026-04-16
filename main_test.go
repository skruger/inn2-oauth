package main

import (
	"bytes"
	"testing"
)

func TestReadAuthRequest_HappyPath(t *testing.T) {
	input := "ClientAuthname: user\r\nClientPassword: pass\r\n.\r\n"
	r := bytes.NewBufferString(input)
	name, pass, err := ReadAuthRequest(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "user" {
		t.Fatalf("expected name 'user', got %q", name)
	}
	if pass != "pass" {
		t.Fatalf("expected pass 'pass', got %q", pass)
	}
}

func TestReadAuthRequest_MissingTerminator(t *testing.T) {
	input := "ClientAuthname: user\r\nClientPassword: pass\r\n"
	r := bytes.NewBufferString(input)
	_, _, err := ReadAuthRequest(r)
	if err == nil {
		t.Fatalf("expected error for missing terminator, got nil")
	}
}

func TestReadAuthRequest_MissingField(t *testing.T) {
	input := "ClientAuthname: user\r\n.\r\n"
	r := bytes.NewBufferString(input)
	_, _, err := ReadAuthRequest(r)
	if err == nil {
		t.Fatalf("expected error for missing field, got nil")
	}
}

func TestReadAuthRequest_ExtraUnknownLines(t *testing.T) {
	input := "Foo: bar\r\nClientPassword: p\r\nClientAuthname: a\r\n.\r\n"
	r := bytes.NewBufferString(input)
	name, pass, err := ReadAuthRequest(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "a" || pass != "p" {
		t.Fatalf("expected (a,p), got (%q,%q)", name, pass)
	}

}

func TestReadAuthRequest_LFOnly(t *testing.T) {
	input := "ClientAuthname: u\nClientPassword: v\n.\n"
	r := bytes.NewBufferString(input)
	name, pass, err := ReadAuthRequest(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "u" || pass != "v" {
		t.Fatalf("expected (u,v), got (%q,%q)", name, pass)
	}
}
