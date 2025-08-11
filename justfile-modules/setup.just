# 🚀 SETUP MODULE  
# Importable justfile module for project setup and initialization
# Usage: import "path/to/setup.just"

# Default paths (can be overridden)
TEMPLATE_ROOT := env_var_or_default("ARCH_LINT_ROOT", justfile_directory())
CONFIG_TEMPLATES := TEMPLATE_ROOT + "/configs/templates"

# Install all tools
install-all-tools:
    @echo "🚀 Installing all linting tools..."
    @just install-arch-tools
    @just install-quality-tools
    @echo "✅ All tools installed successfully!"

# Setup architecture linting for different patterns
setup-arch pattern="clean":
    @echo "📋 Setting up {{pattern}} architecture pattern..."
    @just _copy-arch-config {{pattern}}
    @echo "✅ Architecture setup completed for {{pattern}} pattern!"

# Internal: Copy architecture configuration
_copy-arch-config pattern:
    @if [ -f "{{CONFIG_TEMPLATES}}/.go-arch-lint.{{pattern}}.yml" ]; then \
        cp "{{CONFIG_TEMPLATES}}/.go-arch-lint.{{pattern}}.yml" .go-arch-lint.yml; \
        echo "✅ Copied {{pattern}} architecture config"; \
    elif [ -f "{{CONFIG_TEMPLATES}}/.go-arch-lint.template.yml" ]; then \
        cp "{{CONFIG_TEMPLATES}}/.go-arch-lint.template.yml" .go-arch-lint.yml; \
        echo "✅ Copied template architecture config"; \
    else \
        echo "⚠️ No template found for {{pattern}}, using default"; \
    fi

# Setup code quality configuration
setup-quality strictness="standard":
    @echo "📋 Setting up {{strictness}} quality configuration..."
    @just _copy-quality-config {{strictness}}
    @echo "✅ Quality setup completed for {{strictness}} strictness!"

# Internal: Copy quality configuration
_copy-quality-config strictness:
    @if [ -f "{{CONFIG_TEMPLATES}}/.golangci.{{strictness}}.yml" ]; then \
        cp "{{CONFIG_TEMPLATES}}/.golangci.{{strictness}}.yml" .golangci.yml; \
        echo "✅ Copied {{strictness}} quality config"; \
    elif [ -f "{{CONFIG_TEMPLATES}}/.golangci.template.yml" ]; then \
        cp "{{CONFIG_TEMPLATES}}/.golangci.template.yml" .golangci.yml; \
        echo "✅ Copied template quality config"; \
    else \
        echo "⚠️ No template found for {{strictness}}, using default"; \
    fi

# Complete project setup
setup-project pattern="clean" strictness="standard":
    @echo "🚀 Complete project setup ({{pattern}} architecture, {{strictness}} quality)..."
    @just install-all-tools
    @just setup-arch {{pattern}}
    @just setup-quality {{strictness}}
    @just verify-setup
    @echo "🎉 Project setup completed successfully!"

# Verify complete setup
verify-setup:
    @echo "🔍 Verifying complete setup..."
    @just verify-arch-setup
    @just verify-quality-setup
    @echo "✅ Setup verification completed!"

# Show available templates
list-templates:
    @echo "🗂️ Available architecture patterns:"
    @find "{{CONFIG_TEMPLATES}}" -name ".go-arch-lint.*.yml" -exec basename {} \; 2>/dev/null | sed 's/.go-arch-lint./  - /' | sed 's/.yml//' || echo "  (No templates found)"
    @echo ""
    @echo "🗂️ Available quality configurations:"  
    @find "{{CONFIG_TEMPLATES}}" -name ".golangci.*.yml" -exec basename {} \; 2>/dev/null | sed 's/.golangci./  - /' | sed 's/.yml//' || echo "  (No templates found)"