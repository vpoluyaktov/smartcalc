package calc

import (
	"smartcalc/internal/jwt"
	"strings"
	"testing"
)

func TestJWTReEvaluation(t *testing.T) {
	// Test the actual token that causes the issue
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlFVRXdSVGhFTVRRNE5qVkdOakJFTWtJMk1FWkZRelF5UXpoQk56ZEdNMEpEUlRjeVJrVTBNUSJ9.eyJodHRwczovL3BldGFieXRlLnZldC9lbWFpbCI6InZwb2x1eWFrdG9AY2hld3kuY29tIiwiaXNzIjoiaHR0cHM6Ly9hdXRoLnJoYXBzb2R5LnZldC8iLCJzdWIiOiJva3RhfGNoZXd5LW9rdGEtd29ya2ZvcmNlLXNzby1jb25uZWN0aW9ufDAwdXhuc2xyOG9janpKSUhjMHg3IiwiYXVkIjpbImh0dHBzOi8vcGV0YWJ5dGUudmV0L2FwaS92MS8iLCJodHRwczovL3BldGEtYnl0ZS5hdXRoMC5jb20vdXNlcmluZm8iXSwiaWF0IjoxNzY2NTEyNDg2LCJleHAiOjE3Njc3MjIwODYsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwiLCJhenAiOiJ4NDJxQmJlT3R3UjZkRjN6NjdxRDhJRnlCR2RyZzFFQSJ9.kXD2KxOBbRvDExHuwDwIYJTzZdtgwq7-SX9DkqQYT2YZX8V3U4Jfp-lisK_BxAZBEHJJXhNq2GVEI6637meTlSEzVYMgNUMbyqhaypzeOXY5pNxJTQQ542yOGvB-rnlki6iVkr8FTFY0lPY2FqFF7rbKvcV3CQ2EMblM4Li_BkTpaA_VS0n_m8XV-07404ivjp4e5g65N9H8zXSIgO9za3jAr9VZlQiPtgiVcyGPwQ1P-9GZTBPEM07pDPZkj15BqwAanFhGkFfgddBhtT78dCN1nCpYwdlDKhkZaKKAuAJPpFTtB2QPcXKKf9iS8fpbI8V4G86xn9ChVP6lbcstUA"

	// First evaluation
	lines := []string{
		"# Welcome to SmartCalc!",
		"# Check out the Snippets menu to explore features.",
		"# Type an expression and press Enter to calculate.",
		"",
		"jwt decode " + token + " =",
	}

	t.Log("--- First evaluation ---")
	results := EvalLines(lines, 0)

	// Check that JWT was evaluated successfully
	if !results[4].HasResult {
		t.Fatal("JWT expression should have a result")
	}
	if strings.Contains(results[4].Output, "ERR") {
		t.Fatalf("First evaluation should not have ERR: %s", results[4].Output[:min(100, len(results[4].Output))])
	}
	if !strings.Contains(results[4].Output, "Header:") {
		t.Fatal("JWT output should contain 'Header:'")
	}

	t.Logf("First evaluation output length: %d", len(results[4].Output))
	t.Logf("First evaluation output contains newlines: %v", strings.Contains(results[4].Output, "\n"))

	// Build the new content from results (simulating what frontend does)
	var newLines []string
	for _, r := range results {
		outputLines := strings.Split(r.Output, "\n")
		newLines = append(newLines, outputLines...)
	}

	t.Logf("After first evaluation, total lines: %d", len(newLines))

	// Print the lines to see what we're working with
	for i, line := range newLines {
		if len(line) > 100 {
			t.Logf("Line %d (len=%d): %s...", i+1, len(line), line[:100])
		} else {
			t.Logf("Line %d (len=%d): %s", i+1, len(line), line)
		}
	}

	// Check the JWT line specifically
	jwtLine := newLines[4]
	t.Logf("JWT line ends with: ...%s", jwtLine[len(jwtLine)-20:])
	eq := findResultEquals(jwtLine)
	t.Logf("findResultEquals returned: %d (line length: %d)", eq, len(jwtLine))
	if eq >= 0 {
		expr := jwtLine[:eq]
		t.Logf("Expression ends with: ...%s", expr[len(expr)-20:])
	}

	// Now simulate re-evaluation (what happens when clicking elsewhere)
	t.Log("--- Re-evaluation ---")

	// First, let's see what cleanOutputLines does
	cleanedLines := cleanOutputLines(newLines)
	t.Logf("After cleanOutputLines, total lines: %d", len(cleanedLines))
	for i, line := range cleanedLines {
		if len(line) > 100 {
			t.Logf("Cleaned Line %d (len=%d): %s...", i+1, len(line), line[:100])
		} else {
			t.Logf("Cleaned Line %d (len=%d): %s", i+1, len(line), line)
		}
	}

	// Check the cleaned JWT line
	cleanedJWTLine := cleanedLines[4]
	t.Logf("Cleaned JWT line ends with: ...%s", cleanedJWTLine[len(cleanedJWTLine)-30:])
	cleanedEq := findResultEquals(cleanedJWTLine)
	t.Logf("findResultEquals on cleaned line returned: %d", cleanedEq)

	// Extract the expression and test JWT evaluation directly
	if cleanedEq >= 0 {
		expr := strings.TrimSpace(cleanedJWTLine[:cleanedEq])
		t.Logf("Extracted expression length: %d", len(expr))
		t.Logf("Extracted expression ends with: ...%s", expr[len(expr)-30:])

		// Test IsJWTExpression
		isJWT := jwt.IsJWTExpression(expr)
		t.Logf("IsJWTExpression: %v", isJWT)

		// Test EvalJWT
		jwtResult, err := jwt.EvalJWT(expr)
		if err != nil {
			t.Logf("EvalJWT error: %v", err)
		} else {
			t.Logf("EvalJWT success, result length: %d", len(jwtResult))
		}
	}

	results2 := EvalLines(newLines, 0)

	// Find the JWT expression line in results2
	jwtLineFound := false
	for i, r := range results2 {
		if strings.HasPrefix(cleanedLines[i], "jwt decode") {
			jwtLineFound = true
			t.Logf("JWT line %d: HasResult=%v, Output=%s", i+1, r.HasResult, r.Output[:min(150, len(r.Output))])
			if strings.Contains(r.Output, "ERR") {
				t.Errorf("Re-evaluation should not have ERR: %s", r.Output[:min(100, len(r.Output))])
			}
			if !r.HasResult {
				t.Error("JWT expression should have a result on re-evaluation")
			}
			break
		}
	}

	if !jwtLineFound {
		t.Error("JWT line not found in re-evaluation results")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
