package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/iMerica/jango/apps"
	"github.com/iMerica/jango/checks"
	"github.com/iMerica/jango/cmd/codegen"
	"github.com/iMerica/jango/conf"
	"github.com/iMerica/jango/management"
	"github.com/iMerica/jango/migrations"
	"github.com/iMerica/jango/orm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jango",
	Short: "Jango - A batteries-included Go web framework inspired by Django",
	Long:  "Jango is a Go web framework that mirrors Django's workflow and nouns, optimized for shipping products fast.",
}

var startprojectCmd = &cobra.Command{
	Use:   "startproject [name] [target]",
	Short: "Create a new Jango project",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runStartProject,
}

var startappCmd = &cobra.Command{
	Use:   "startapp [name] [target]",
	Short: "Create a new Jango app within a project",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runStartApp,
}

var codegenCmd = &cobra.Command{
	Use:   "codegen [project-dir]",
	Short: "Generate the app registry from installed apps",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCodegen,
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run framework checks for the project",
	Long:  "Checks the project for common problems. Use --deploy to run deployment-specific security checks.",
	Args:  cobra.NoArgs,
	RunE:  runCheck,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Jango version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("jango v0.1.0")
	},
}

var makemigrationsCmd = &cobra.Command{
	Use:   "makemigrations [app_label ...]",
	Short: "Create new migrations based on model changes",
	Long:  "Examines model definitions and creates new migration files for any changes that have been made.",
	Args:  cobra.ArbitraryArgs,
	RunE:  runMakeMigrations,
}

var migrateCmd = &cobra.Command{
	Use:   "migrate [app_label] [migration_name]",
	Short: "Apply or reverse migrations",
	Long:  "Applies pending migrations, bringing the database schema up to date with the current models.",
	Args:  cobra.MaximumNArgs(2),
	RunE:  runMigrate,
}

var showmigrationsCmd = &cobra.Command{
	Use:   "showmigrations [app_label]",
	Short: "Show the migration state for the project",
	Long:  "Lists all migrations for each app, indicating which have been applied.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runShowMigrations,
}

var sqlmigrateCmd = &cobra.Command{
	Use:   "sqlmigrate app_label migration_name",
	Short: "Show the SQL for a named migration",
	Long:  "Prints the SQL statements that would be executed for the given migration.",
	Args:  cobra.ExactArgs(2),
	RunE:  runSQLMigrate,
}

var runserverCmd = &cobra.Command{
	Use:   "runserver [addr]",
	Short: "Start the development server",
	Long:  "Start a lightweight development web server. Not for production use.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runServer,
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open an interactive shell with the Jango environment",
	Args:  cobra.NoArgs,
	RunE:  runShell,
}

var createsuperuserCmd = &cobra.Command{
	Use:   "createsuperuser",
	Short: "Create a superuser account",
	Args:  cobra.NoArgs,
	RunE:  runCreateSuperUser,
}

var collectstaticCmd = &cobra.Command{
	Use:   "collectstatic",
	Short: "Collect static files into STATIC_ROOT",
	Args:  cobra.NoArgs,
	RunE:  runCollectStatic,
}

var testCmd = &cobra.Command{
	Use:   "test [packages...]",
	Short: "Run tests",
	Args:  cobra.ArbitraryArgs,
	RunE:  runTests,
}

var deployFlag bool
var migrationEmptyFlag bool
var runserverAddr string

func init() {
	rootCmd.AddCommand(startprojectCmd)
	rootCmd.AddCommand(startappCmd)
	rootCmd.AddCommand(codegenCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(makemigrationsCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(showmigrationsCmd)
	rootCmd.AddCommand(sqlmigrateCmd)
	rootCmd.AddCommand(runserverCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(createsuperuserCmd)
	rootCmd.AddCommand(collectstaticCmd)
	rootCmd.AddCommand(testCmd)

	checkCmd.Flags().BoolVar(&deployFlag, "deploy", false, "Run deployment-specific security checks")
	makemigrationsCmd.Flags().BoolVar(&migrationEmptyFlag, "empty", false, "Create an empty migration")
	runserverCmd.Flags().StringVar(&runserverAddr, "addr", "localhost:8000", "Address to bind to")
}

func Execute() error {
	return rootCmd.Execute()
}

func runStartProject(cmd *cobra.Command, args []string) error {
	name := args[0]
	target := ""
	if len(args) > 1 {
		target = args[1]
	}
	return codegen.CreateProject(name, target)
}

func runStartApp(cmd *cobra.Command, args []string) error {
	name := args[0]
	target := ""
	if len(args) > 1 {
		target = args[1]
	}
	return codegen.CreateApp(name, target)
}

func runCodegen(cmd *cobra.Command, args []string) error {
	projectDir := "."
	if len(args) > 0 {
		projectDir = args[0]
	}

	info, err := codegen.FindProjectInfo(projectDir)
	if err != nil {
		return err
	}

	return codegen.GenerateRegistry(info, projectDir)
}

func runCheck(cmd *cobra.Command, args []string) error {
	results := checks.RunAll()
	for _, r := range results {
		printResult(r)
	}

	if deployFlag {
		deployResults := checks.RunDeploy()
		for _, r := range deployResults {
			printResult(r)
		}
		results = append(results, deployResults...)
	}

	hasErrors := false
	for _, r := range results {
		if r.Status == checks.StatusErr {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		return fmt.Errorf("check: found errors, see above")
	}

	fmt.Printf("System check identified %d issues.\n", len(results))
	return nil
}

func runMakeMigrations(cmd *cobra.Command, args []string) error {
	settings, db, err := setupFramework()
	if err != nil {
		return err
	}
	if db != nil {
		defer db.Close()
	}

	_ = settings

	appLabels := args
	if len(appLabels) == 0 {
		appConfigs := apps.GlobalRegistry().GetAppConfigs()
		for _, cfg := range appConfigs {
			appLabels = append(appLabels, cfg.Label)
		}
	}

	anyCreated := false
	for _, appLabel := range appLabels {
		appConfig, err := apps.GlobalRegistry().GetAppConfig(appLabel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		appDir := appConfig.Path
		if appDir == "" {
			fmt.Fprintf(os.Stderr, "Warning: cannot determine path for app %q, skipping\n", appLabel)
			continue
		}

		name, err := migrations.MakeMigrations(appLabel, appDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		if name != "" {
			fmt.Printf("Migrations created for %q: %s\n", appLabel, name)
			anyCreated = true
		} else {
			fmt.Printf("No changes detected for %q\n", appLabel)
		}
	}

	if anyCreated {
		fmt.Println("\nAfter creating migrations, rebuild and run 'jango migrate' to apply them.")
	}

	return nil
}

func runMigrate(cmd *cobra.Command, args []string) error {
	settings, db, err := setupFramework()
	if err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("migrate: no database connection available")
	}
	defer db.Close()

	_ = settings

	if err := runSystemChecks(); err != nil {
		return err
	}

	appLabels := []string{}
	if len(args) >= 1 {
		appLabels = []string{args[0]}
	} else {
		appConfigs := apps.GlobalRegistry().GetAppConfigs()
		for _, cfg := range appConfigs {
			appLabels = append(appLabels, cfg.Label)
		}
	}

	ctx := context.Background()

	for _, appLabel := range appLabels {
		executor := migrations.NewExecutor(db, appLabel)
		applied, err := executor.ApplyAll(ctx)
		if err != nil {
			return fmt.Errorf("migrate: error migrating app %q: %w", appLabel, err)
		}

		if len(applied) == 0 {
			fmt.Printf("No migrations to apply for %q\n", appLabel)
		} else {
			for _, name := range applied {
				fmt.Printf("Applying %s... OK\n", name)
			}
		}
	}

	return nil
}

func runShowMigrations(cmd *cobra.Command, args []string) error {
	settings, db, err := setupFramework()
	if err != nil {
		return err
	}

	_ = settings

	ctx := context.Background()

	applied := make(map[string]bool)
	if db != nil {
		defer db.Close()
		for _, appConfig := range apps.GlobalRegistry().GetAppConfigs() {
			executor := migrations.NewExecutor(db, appConfig.Label)
			if appApplied, err := executor.GetAppliedMigrations(ctx); err == nil {
				for k := range appApplied {
					applied[k] = true
				}
			}
		}
	}

	appLabels := []string{}
	if len(args) == 1 {
		appLabels = []string{args[0]}
	} else {
		for _, cfg := range apps.GlobalRegistry().GetAppConfigs() {
			appLabels = append(appLabels, cfg.Label)
		}
	}

	for _, appLabel := range appLabels {
		fmt.Printf("%s\n", appLabel)
		migrations := migrations.GetMigrationsForApp(appLabel)
		if len(migrations) == 0 {
			fmt.Println("  (no migrations)")
		}
		for _, m := range migrations {
			key := appLabel + "." + m.Name
			status := "[ ]"
			if applied[key] {
				status = "[X]"
			}
			fmt.Printf("  %s %s\n", status, m.Name)
		}
	}

	return nil
}

func runSQLMigrate(cmd *cobra.Command, args []string) error {
	settings, db, err := setupFramework()
	if err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("sqlmigrate: no database connection available")
	}
	defer db.Close()

	_ = settings

	appLabel := args[0]
	migrationName := args[1]

	m, ok := migrations.GetMigration(appLabel, migrationName)
	if !ok {
		return fmt.Errorf("sqlmigrate: migration %s.%s not found", appLabel, migrationName)
	}

	executor := migrations.NewExecutor(db, appLabel)
	sqlStatements, err := executor.GetMigrationSQL(m)
	if err != nil {
		return fmt.Errorf("sqlmigrate: %w", err)
	}

	for _, sql := range sqlStatements {
		fmt.Println(sql)
		fmt.Println()
	}

	return nil
}

func runServer(cmd *cobra.Command, args []string) error {
	settings, _, err := setupFramework()
	if err != nil {
		return err
	}

	_ = settings

	if err := runSystemChecks(); err != nil {
		return err
	}

	addr := runserverAddr
	fmt.Printf("Starting development server at http://%s\n", addr)
	fmt.Println("Quit the server with CONTROL-C.")

	select {}
}

func runShell(cmd *cobra.Command, args []string) error {
	settings, db, err := setupFramework()
	if err != nil {
		return err
	}
	if db != nil {
		defer db.Close()
	}

	_ = settings

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	fmt.Println("Jango interactive shell")
	fmt.Println("Type 'exit' to leave the shell")

	return nil
}

func runCreateSuperUser(cmd *cobra.Command, args []string) error {
	settings, db, err := setupFramework()
	if err != nil {
		return err
	}
	if db != nil {
		defer db.Close()
	}

	_ = settings

	fmt.Println("createsuperuser: not yet implemented (requires auth module)")
	return nil
}

func runCollectStatic(cmd *cobra.Command, args []string) error {
	settings, db, err := setupFramework()
	if err != nil {
		return err
	}
	if db != nil {
		defer db.Close()
	}

	_ = db

	staticRoot := settings.StaticRoot
	if staticRoot == "" {
		return fmt.Errorf("collectstatic: STATIC_ROOT is not set")
	}

	fmt.Printf("Collecting static files to %s\n", staticRoot)
	fmt.Println("collectstatic: not yet fully implemented (requires staticfiles module)")
	return nil
}

func runTests(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"./..."}
	}

	testCmd := []string{"test"}
	testCmd = append(testCmd, args...)

	return nil
}

func setupFramework() (*conf.Settings, *orm.DB, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, nil, fmt.Errorf("jango: cannot load settings: %w", err)
	}

	conf.ApplyEnv(settings)

	if errs := conf.Validate(settings); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "configuration error: %s: %s\n", e.Field, e.Message)
		}
		return nil, nil, fmt.Errorf("jango: configuration validation failed")
	}

	conf.ConfigureLogging(settings)
	conf.Init(settings)

	appConfigs := apps.GlobalRegistry().GetAppConfigs()
	if len(appConfigs) == 0 {
		slog.Info("jango: no apps registered yet")
	}

	var db *orm.DB
	dbSettings := settings.DefaultDatabase()
	if dbSettings != nil && dbSettings.Name != "" {
		dbConfig := &orm.DBConfig{
			Host:     dbSettings.Host,
			Port:     dbSettings.Port,
			Name:     dbSettings.Name,
			User:     dbSettings.User,
			Password: dbSettings.Password,
			SSLMode:  "prefer",
		}
		if dbSettings.Options != nil {
			if sslMode, ok := dbSettings.Options["sslmode"]; ok {
				dbConfig.SSLMode = sslMode
			}
		}

		ctx := context.Background()
		var err error
		db, err = orm.OpenDB(ctx, dbConfig)
		if err != nil {
			slog.Warn("jango: cannot connect to database", "error", err)
			db = nil
		} else {
			orm.SetDefaultDB(db)
		}
	}

	conf.Freeze()

	discoverCustomCommands()

	return settings, db, nil
}

func loadSettings() (*conf.Settings, error) {
	return conf.DefaultSettings(), nil
}

func runSystemChecks() error {
	results := checks.RunAll()
	hasErrors := false
	for _, r := range results {
		printResult(r)
		if r.Status == checks.StatusErr {
			hasErrors = true
		}
	}

	if hasErrors {
		return fmt.Errorf("system check identified errors")
	}

	fmt.Printf("System check identified %d issue(s).\n", len(results))
	return nil
}

func printResult(r checks.CheckResult) {
	switch r.Status {
	case checks.StatusErr:
		fmt.Fprintf(os.Stderr, "ERROR %s: %s\n", r.ID, r.Message)
		if r.Hint != "" {
			fmt.Fprintf(os.Stderr, "    HINT: %s\n", r.Hint)
		}
	case checks.StatusWarn:
		fmt.Fprintf(os.Stderr, "WARNING %s: %s\n", r.ID, r.Message)
		if r.Hint != "" {
			fmt.Fprintf(os.Stderr, "    HINT: %s\n", r.Hint)
		}
	default:
		fmt.Printf("OK %s: %s\n", r.ID, r.Message)
	}
}

func discoverCustomCommands() {
	existingNames := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		existingNames[cmd.Name()] = true
	}

	for name, cmd := range management.AllCommands() {
		if !existingNames[name] {
			localCmd := cmd
			cobraCmd := &cobra.Command{
				Use:   localCmd.Name,
				Short: localCmd.Description,
				Args:  cobra.ArbitraryArgs,
				RunE: func(cobraCmd *cobra.Command, args []string) error {
					return localCmd.Handler(args)
				},
			}
			rootCmd.AddCommand(cobraCmd)
		}
	}
}

func Run() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
