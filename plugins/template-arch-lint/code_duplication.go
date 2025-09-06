package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// CodeBlock represents a block of code for duplication analysis
type CodeBlock struct {
	Node     ast.Node
	Hash     string
	Tokens   int
	Filename string
	StartPos token.Pos
	EndPos   token.Pos
}

// runCodeDuplicationDetection implements code duplication detection analyzer
func runCodeDuplicationDetection(pass *analysis.Pass) (interface{}, error) {
	const minTokens = 15 // Configurable threshold

	var codeBlocks []CodeBlock

	// Extract code blocks from all files
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		blocks := extractCodeBlocks(pass, file, filename, minTokens)
		codeBlocks = append(codeBlocks, blocks...)
	}

	// Find duplicates
	duplicates := findDuplicateBlocks(codeBlocks)

	// Report duplicates
	for _, group := range duplicates {
		if len(group) > 1 {
			first := group[0]
			pass.Reportf(first.StartPos,
				"CODE_DUPLICATION: Duplicated code block (%d tokens) found in %d locations. Consider extracting to a function.",
				first.Tokens, len(group))

			for i, block := range group[1:] {
				if i < 3 { // Limit to first 3 duplicates to avoid spam
					pass.Reportf(block.StartPos,
						"CODE_DUPLICATION: Duplicate of code at %s",
						pass.Fset.Position(first.StartPos))
				}
			}
		}
	}

	return nil, nil
}

// extractCodeBlocks extracts analyzable code blocks from a file
func extractCodeBlocks(pass *analysis.Pass, file *ast.File, filename string, minTokens int) []CodeBlock {
	var blocks []CodeBlock

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body != nil && len(node.Body.List) >= 3 {
				block := createCodeBlock(pass, node.Body, filename, minTokens)
				if block != nil {
					blocks = append(blocks, *block)
				}
			}
		case *ast.BlockStmt:
			if len(node.List) >= 3 {
				block := createCodeBlock(pass, node, filename, minTokens)
				if block != nil {
					blocks = append(blocks, *block)
				}
			}
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			// Analyze complex statements
			block := createCodeBlock(pass, node, filename, minTokens)
			if block != nil {
				blocks = append(blocks, *block)
			}
		}
		return true
	})

	return blocks
}

// createCodeBlock creates a code block for duplication analysis
func createCodeBlock(pass *analysis.Pass, node ast.Node, filename string, minTokens int) *CodeBlock {
	if node == nil {
		return nil
	}

	// Calculate approximate token count
	tokenCount := estimateTokenCount(node)
	if tokenCount < minTokens {
		return nil
	}

	// Generate a structural hash (simplified)
	hash := generateStructuralHash(node)

	return &CodeBlock{
		Node:     node,
		Hash:     hash,
		Tokens:   tokenCount,
		Filename: filename,
		StartPos: node.Pos(),
		EndPos:   node.End(),
	}
}

// estimateTokenCount provides a rough estimate of tokens in an AST node
func estimateTokenCount(node ast.Node) int {
	count := 0
	ast.Inspect(node, func(n ast.Node) bool {
		if n != nil {
			count++
		}
		return true
	})
	return count
}

// generateStructuralHash creates a hash based on AST structure (simplified)
func generateStructuralHash(node ast.Node) string {
	var builder strings.Builder

	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		// Include node type in hash
		nodeType := reflect.TypeOf(n).String()
		builder.WriteString(nodeType)
		builder.WriteString(";")

		// Include specific patterns for different node types
		switch typed := n.(type) {
		case *ast.Ident:
			// Don't include actual identifier names, just that it's an identifier
			builder.WriteString("IDENT;")
		case *ast.BasicLit:
			// Include literal type but not value
			builder.WriteString(fmt.Sprintf("LIT_%s;", typed.Kind.String()))
		case *ast.BinaryExpr:
			// Include operator
			builder.WriteString(fmt.Sprintf("BINOP_%s;", typed.Op.String()))
		case *ast.UnaryExpr:
			// Include operator
			builder.WriteString(fmt.Sprintf("UNOP_%s;", typed.Op.String()))
		}

		return true
	})

	return builder.String()
}

// findDuplicateBlocks groups code blocks by their structural similarity
func findDuplicateBlocks(blocks []CodeBlock) [][]CodeBlock {
	hashGroups := make(map[string][]CodeBlock)

	// Group by hash
	for _, block := range blocks {
		hashGroups[block.Hash] = append(hashGroups[block.Hash], block)
	}

	// Return only groups with duplicates
	var duplicates [][]CodeBlock
	for _, group := range hashGroups {
		if len(group) > 1 {
			// Additional similarity check to reduce false positives
			if areSimilarBlocks(group) {
				duplicates = append(duplicates, group)
			}
		}
	}

	return duplicates
}

// areSimilarBlocks performs additional similarity checks beyond hash matching
func areSimilarBlocks(blocks []CodeBlock) bool {
	if len(blocks) < 2 {
		return false
	}

	first := blocks[0]

	for _, block := range blocks[1:] {
		// Check token count similarity (within 20% difference)
		tokenDiff := abs(first.Tokens - block.Tokens)
		maxTokens := max(first.Tokens, block.Tokens)

		if float64(tokenDiff)/float64(maxTokens) > 0.2 {
			return false
		}
	}

	return true
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
