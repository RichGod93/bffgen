package commands

import (
	"fmt"
	"strings"

	"github.com/RichGod93/bffgen/internal/templates"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Template management",
	Long:  `Manage project templates - list, show details, and install community templates`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `List all available project templates including built-in and community templates`,
	RunE:  runTemplateList,
}

var templateShowCmd = &cobra.Command{
	Use:   "show <template-name>",
	Short: "Show template details",
	Long:  `Display detailed information about a specific template`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateShow,
}

var templateInstallCmd = &cobra.Command{
	Use:   "install <git-url>",
	Short: "Install template from Git repository",
	Long:  `Install a community template from a Git repository`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateInstall,
}

var templateRemoveCmd = &cobra.Command{
	Use:   "remove <template-name>",
	Short: "Remove an installed template",
	Long:  `Remove a community template from local storage`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateRemove,
}

var templateUpdateCmd = &cobra.Command{
	Use:   "update <template-name>",
	Short: "Update an installed template",
	Long:  `Update a template to the latest version from its source`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateUpdate,
}

func init() {
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateInstallCmd)
	templateCmd.AddCommand(templateRemoveCmd)
	templateCmd.AddCommand(templateUpdateCmd)
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	templatesDir, err := templates.GetDefaultTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	manager := templates.NewManager(templatesDir)

	// Ensure templates directory exists
	if err := manager.EnsureTemplatesDir(); err != nil {
		return err
	}

	templateList, err := manager.List()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templateList) == 0 {
		fmt.Println("No templates available.")
		fmt.Println("\nTo get started, templates will be bundled in future releases.")
		return nil
	}

	fmt.Println("\nAvailable Templates:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	for _, tmpl := range templateList {
		fmt.Printf("📦 %s (v%s)\n", tmpl.Name, tmpl.Version)
		fmt.Printf("   %s\n", tmpl.Description)
		fmt.Printf("   Language: %s\n", tmpl.Language)

		if len(tmpl.Features) > 0 {
			fmt.Printf("   Features: %s\n", strings.Join(tmpl.Features, ", "))
		}

		fmt.Println()
	}

	return nil
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	templatesDir, err := templates.GetDefaultTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	manager := templates.NewManager(templatesDir)

	tmpl, err := manager.Get(templateName)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	fmt.Printf("\nTemplate: %s\n", tmpl.Name)
	fmt.Printf("Version: %s\n", tmpl.Version)
	fmt.Printf("Author: %s\n", tmpl.Author)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	fmt.Println("Description:")
	fmt.Printf("  %s\n\n", tmpl.Description)

	if len(tmpl.Features) > 0 {
		fmt.Println("Features:")
		for _, feature := range tmpl.Features {
			fmt.Printf("  ✓ %s\n", feature)
		}
		fmt.Println()
	}

	requiredVars := tmpl.GetRequiredVariables()
	if len(requiredVars) > 0 {
		fmt.Println("Required Variables:")
		for _, v := range requiredVars {
			fmt.Printf("  %s - %s\n", v.Name, v.Description)
		}
		fmt.Println()
	}

	fmt.Printf("Usage:\n")
	fmt.Printf("  bffgen init my-project --template %s\n\n", tmpl.Name)

	return nil
}

func runTemplateInstall(cmd *cobra.Command, args []string) error {
	gitURL := args[0]

	templatesDir, err := templates.GetDefaultTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	manager := templates.NewManager(templatesDir)
	if err := manager.EnsureTemplatesDir(); err != nil {
		return err
	}

	installer := templates.NewInstaller(templatesDir)

	// Install from GitHub
	template, err := installer.InstallFromGitHub(gitURL)
	if err != nil {
		return err
	}

	// Show template info
	fmt.Println()
	fmt.Printf("📦 %s (v%s)\n", template.Name, template.Version)
	fmt.Printf("   %s\n", template.Description)

	if len(template.Features) > 0 {
		fmt.Printf("   Features: %s\n", strings.Join(template.Features, ", "))
	}

	fmt.Println()
	fmt.Printf("💡 Usage: bffgen init my-project --template %s\n", template.Name)

	return nil
}

func runTemplateRemove(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	templatesDir, err := templates.GetDefaultTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	installer := templates.NewInstaller(templatesDir)
	return installer.Remove(templateName)
}

func runTemplateUpdate(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	templatesDir, err := templates.GetDefaultTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	installer := templates.NewInstaller(templatesDir)
	return installer.Update(templateName)
}

var templateRegistryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage template registry",
	Long:  `Search and update the community template registry`,
}

var templateRegistryUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update template registry",
	RunE:  runTemplateRegistryUpdate,
}

var templateRegistryListCmd = &cobra.Command{
	Use:   "list [term]",
	Short: "List or search registry templates",
	RunE:  runTemplateRegistryList,
}

func init() {
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateInstallCmd)

	templateCmd.AddCommand(templateRegistryCmd)
	templateRegistryCmd.AddCommand(templateRegistryUpdateCmd)
	templateRegistryCmd.AddCommand(templateRegistryListCmd)
}

func runTemplateRegistryUpdate(cmd *cobra.Command, args []string) error {
	templatesDir, err := templates.GetDefaultTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	manager := templates.NewManager(templatesDir)
	if err := manager.EnsureTemplatesDir(); err != nil {
		return err
	}

	fmt.Println("🔄 Updating template registry...")
	if err := manager.UpdateRegistry(); err != nil {
		return fmt.Errorf("failed to update registry: %w", err)
	}

	fmt.Println("✅ Registry updated successfully")
	return nil
}

func runTemplateRegistryList(cmd *cobra.Command, args []string) error {
	templatesDir, err := templates.GetDefaultTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	manager := templates.NewManager(templatesDir)
	registry, err := manager.GetRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Filter if search term provided
	var entries []templates.RegistryEntry
	searchTerm := ""
	if len(args) > 0 {
		searchTerm = strings.ToLower(args[0])
	}

	for _, entry := range registry.Templates {
		if searchTerm == "" ||
			strings.Contains(strings.ToLower(entry.Name), searchTerm) ||
			strings.Contains(strings.ToLower(entry.Description), searchTerm) {
			entries = append(entries, entry)
		}
	}

	if len(entries) == 0 {
		if searchTerm != "" {
			fmt.Printf("No templates found matching '%s'\n", searchTerm)
		} else {
			fmt.Println("Registry is empty or not yet updated. Run 'bffgen template registry update'.")
		}
		return nil
	}

	fmt.Println("\nCommunity Registry Templates:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	for _, entry := range entries {
		fmt.Printf("📦 %s (v%s)\n", entry.Name, entry.Version)
		fmt.Printf("   %s\n", entry.Description)
		fmt.Printf("   Language: %s\n", entry.Language)
		fmt.Printf("   URL: %s\n", entry.URL)
		fmt.Println()
	}

	return nil
}
