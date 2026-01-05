//┌───────────────────────────────────────────────────────────────────────────┐
//  File           aay.go
//  Description    An info viewer and launcher for installed android packages
//  Version        0.1.0 alpha
//  Author         Moin Khan <m33mt33n>
//  License        GNU General Public License v3.0 or later (see LICENSE)
//  Created        June 22, 2020, 02:27
//  Re-Written     November 25, 2025 00:44 (translated from python with updates)
//  Last Updated   January 06, 2026 01:14
//└───────────────────────────────────────────────────────────────────────────┘

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/google/shlex"
	"github.com/m33mt33n/apkg"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	prog       string = "aay"
	version    string = "0.1.0 alpha [06.01.2026]"
	debug_mode bool   = false
)

var (
	rootcmd        *cobra.Command
	a_show_version bool
)

func exit_on_error1(rcode uint8) {
	if rcode != 0 {
		fmt.Fprintf(os.Stderr, apkg.Error_msgs[int(rcode)]+"\n", prog)
		os.Exit(int(rcode))
	}
}

func exit_on_error2(rcode int, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", prog, err.Error())
		os.Exit(rcode)
	}
}

func get_action(cmd_usage string) string {
	split := strings.Split(cmd_usage, " ")
	return split[0]
}

func validate_apks(apks []string) {
	for _, apk := range apks {
		message := ""
		if !apkg.File_exists(apk) {
			message = "invalid path"
		} else if !apkg.Valid_apk(apk) {
			message = "invalid apk file"
		}
		if len(message) > 0 {
			fmt.Fprintf(os.Stderr, "error: %s: %s\n", message, apk)
			os.Exit(1)
		}
	}
}

func apk_action(cmd *cobra.Command, args []string) {
	var (
		action = get_action(cmd.Use)
		rcode  = 0
	)
	if action == "_yet_undefined_" {
		// do nothing
	} else {
		apklist := []apkg.ApkInfo{}
		completed_with_errors := false
		apkinfo_json, _ := cmd.Flags().GetBool("json")
		num_of_apks := len(args)
		for idx, apk := range args {
			apkinfo, rcode := apkg.Get_info_from_apk(apk)
			if rcode == 0 {
				if apkinfo_json {
					apklist = append(apklist, apkinfo)
				} else {
					fmt.Println(apkinfo)
					if num_of_apks > 1 && (idx+1) < num_of_apks {
						fmt.Println()
					}
				}
			} else {
				fmt.Printf(apkg.Error_msgs[int(rcode)]+"\n", prog)
				completed_with_errors = true
			}
		}
		if apkinfo_json {
			bytes, err := json.Marshal(apklist)
			exit_on_error2(65, err)
			os.Stdout.Write(bytes)
		}
		rcode = 0
		if completed_with_errors {
			rcode = 66
		}
	}
	os.Exit(int(rcode))
}

func app_action(cmd *cobra.Command, args []string) {
	var (
		pattern string
		regex   *regexp.Regexp
		err     error
	)
	action := get_action(cmd.Use)
	if len(args) == 1 {
		pattern = args[0]
	}
	pattern_is_regex := false
	if cmd.Flags().Changed("regex") {
		pattern_is_regex = true
		if !cmd.Flags().Changed("case-sensitive") {
			pattern = "(?i)" + pattern
		}
		regex, err = regexp.Compile(pattern)
		exit_on_error2(63, err)
	}
	use_pkgname, _ := cmd.Flags().GetBool("use-pkgname")
	appinfo_json, _ := cmd.Flags().GetBool("json")
	applist, rcode := apkg.Get_applist(use_pkgname, pattern, pattern_is_regex, regex)
	exit_on_error1(rcode)
	var app apkg.App
	if len(applist) == 0 {
		exit_on_error1(67)
	} else if len(applist) == 1 {
		app = applist[0]
	} else {
		if appinfo_json {
			exit_on_error1(68)
		} else {
			app, rcode = apkg.Select(applist)
			exit_on_error1(rcode)
		}
	}
	if slices.Contains([]string{"open-datadir", "datadir-path", "apkpath", "uid"}, action) {
		inst_info, rcode := apkg.Get_installation_info(app.Package)
		exit_on_error1(rcode)
		if action == "open-datadir" {
			open_cmd, _ := cmd.Flags().GetString("command")
			if open_cmd == "gui" {
				open_cmd = fmt.Sprintf(
					"am start -a android.intent.action.VIEW -d 'file://%s'", inst_info.Data_dir,
				)
				_, err := exec.Command("sh", "-c", open_cmd).CombinedOutput()
				exit_on_error2(69, err)
			} else {
				if open_cmd == "vifm" {
					open_cmd = "vifm %s"
				}
				command, err := shlex.Split(fmt.Sprintf(open_cmd, inst_info.Data_dir))
				exit_on_error2(70, err)
				var executable string
				executable, err = exec.LookPath(command[0])
				exit_on_error2(71, err)
				if strings.Contains(command[0], "/") {
					command[0] = filepath.Base(command[0])
				}
				err = syscall.Exec(executable, command, os.Environ())
				exit_on_error2(72, err)
			}
		} else if action == "datadir-path" {
			println(inst_info.Data_dir)
		} else if action == "apkpath" {
			println(inst_info.Apk_path)
		} else if action == "uid" {
			println(inst_info.UID)
		}
		os.Exit(0)
	}
	rcode = 0
	if action == "force-stop" {
		rcode = app.Package.Force_stop(debug_mode)
	} else if action == "safe-kill" {
		rcode = app.Package.Kill(true, debug_mode)
	} else if action == "kill" {
		rcode = app.Package.Kill(false, debug_mode)
	} else if action == "enable" {
		rcode = app.Package.Change_state("enable", debug_mode)
	} else if action == "disable" {
		rcode = app.Package.Change_state("disable", debug_mode)
	} else if action == "launch" {
		rcode = app.Start(debug_mode)
	} else {
		if appinfo_json {
			inst_info, rc := apkg.Get_installation_info(app.Package)
			if rc == 0 {
				app.Apk_path = inst_info.Apk_path
				app.Data_dir = inst_info.Data_dir
				app.Library_path = inst_info.Library_path
				app.Resource_path = inst_info.Resource_path
			} else {
				app.Apk_path = "NA"
				app.Data_dir = "NA"
				app.Library_path = "NA"
				app.Resource_path = "NA"
			}
			bytes, err := json.Marshal(app)
			exit_on_error2(73, err)
			os.Stdout.Write(bytes)
		} else {
			fmt.Println(app)
		}
	}
	exit_on_error1(rcode)
	os.Exit(0)
}

func db_action(cmd *cobra.Command, args []string) {
	action := get_action(cmd.Use)
	if action == "update" {
		exit_on_error1(apkg.Create_db(false))
	} else if action == "check" {
		exit_on_error1(apkg.Check_db())
	} else if action == "last-update-time" {
		timestamp, rcode := apkg.Get_last_update_time()
		exit_on_error1(rcode)
		fmt.Printf("Last database update: %s", timestamp.Format(time.DateTime))
	}
	os.Exit(0)
}

func init() {
	var made_a_new_db bool
	var rcode uint8
	if _, err := exec.LookPath("aapt"); err != nil {
		apkg.Aapt_path, rcode = embeded_aapt()
		exit_on_error1(rcode)
	}
	size, _ := apkg.File_size(apkg.DB_path)
	if size <= 0 {
		exit_on_error1(apkg.Create_db(true))
		made_a_new_db = true
	}
	rootcmd = &cobra.Command{
		Use:   "aay",
		Short: "An info viewer and launcher for installed android packages",
		Args:  cobra.NoArgs,
		//Version: version,  // uses `-v` as short name but `-V` is better.
		Run: func(cmd *cobra.Command, args []string) {
			if a_show_version {
				fmt.Printf("%s %s, License: GPLv3+, Author: Moin Khan [m33mt33n]\n", prog, version)
				return
			}
			cmd.Help()
		},
	}
	var mancmd = &cobra.Command{
		Use:   "man",
		Short: "Generate man pages",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			header := &doc.GenManHeader{
				Title:   "aay",
				Section: "1",
			}
			output_dir := "./man"
			if err := os.MkdirAll(output_dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to create man dir: %v\n", err)
				os.Exit(1)
			}
			if err := doc.GenManTree(rootcmd, header, output_dir); err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to generate man pages: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("aay: man pages generated in %s\n", output_dir)
		},
	}
	rootcmd.AddCommand(mancmd)
	rootcmd.Flags().BoolVarP(
		&a_show_version, "version", "V", false, "show version number and exit",
	)

	subcmd_apk := &cobra.Command{
		Use:     "apk",
		Aliases: []string{"p"},
		Short:   "Android package (APK) files related actions (default sub-command is `info` if omitted)",
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				validate_apks(args)
				apk_action(cmd, args)
			} else {
				exit_on_error1(74)
			}
		},
	}
	subcmd_apk.Flags().BoolP(
		"json", "j", false, "output app info in JSON",
	)

	subcmd_app := &cobra.Command{
		Use:     "app",
		Aliases: []string{"a"},
		Short:   "Actions related to installed apps (default sub-command is `info` if omitted)",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			app_action(cmd, args)
		},
	}
	subcmd_app.PersistentFlags().BoolP(
		"case-sensitive", "I", false, "case-sensitive pattern matching",
	)
	subcmd_app.Flags().BoolP(
		"json", "j", false, "output app info in JSON",
	)
	subcmd_app.PersistentFlags().BoolP(
		"regex", "r", false, "treat pattern as regexp",
	)
	subcmd_app.PersistentFlags().BoolP(
		"use-pkgname", "p", false,
		"match package name instead of app name while matching pattern",
	)

	subcmd_db := &cobra.Command{
		Use:     "db",
		Aliases: []string{"d"},
		Short:   "Manage app database",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	for _, cmd := range [3]*cobra.Command{subcmd_app, subcmd_apk, subcmd_db} {
		rootcmd.AddCommand(cmd)
	}

	apk_actions := [1]*cobra.Command{
		&cobra.Command{
			Use:     "info <apk> [apk ...]",
			Aliases: []string{"i"},
			Short:   "get app info from apk file(s)",
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				validate_apks(args)
				apk_action(cmd, args)
			},
		},
	}
	for _, action := range apk_actions {
		if strings.HasPrefix(action.Use, "info") {
			action.Flags().BoolP("json", "j", false, "output app info in JSON")
    }
		subcmd_apk.AddCommand(action)
	}

	app_actions := [11]*cobra.Command{
		&cobra.Command{
			Use:     "apkpath <pattern>",
			Aliases: []string{"a"},
			Short:   "print APK path for the selected app",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "datadir-path <pattern>",
			Aliases: []string{"p"},
			Short:   "print data directory path for the selected app",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "disable <pattern>",
			Aliases: []string{"d"},
			Short:   "disable package",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "enable <pattern>",
			Aliases: []string{"e"},
			Short:   "enable package",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "force-stop <pattern>",
			Aliases: []string{"f"},
			Short:   "force-stop app",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "info <pattern>",
			Aliases: []string{"i"},
			Short:   "print app info",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "kill <pattern>",
			Aliases: []string{"k"},
			Short:   "kill app using `pkill` command",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "launch <pattern>",
			Aliases: []string{"l"},
			Short:   "launch app",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "open-datadir <pattern>",
			Aliases: []string{"o"},
			Short:   "open data directory with provided command",
			Args:    cobra.ExactArgs(1),
			Example: fmt.Sprintf(
				"%s\n%s",
				"aay app open-datadir [-c cmd] <pattern>",
				"aay a o -c gui <pattern>",
			),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "safe-kill <pattern>",
			Aliases: []string{"s"},
			Short:   "kill app using `am kill` command",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "uid <pattern>",
			Aliases: []string{"u"},
			Short:   "print UID for the selected app",
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				app_action(cmd, args)
			},
		},
	}
	for _, action := range app_actions {
		if strings.HasPrefix(action.Use, "open-datadir") {
			action.Flags().StringP(
				"command", "c", "vifm", fmt.Sprintf(
					"%s %s %s %s",
					"command to be used in opening of data directory, could be one of shorthand",
					"commands 'gui' (for using android GUI app) or '`vifm`' (for using vifm, if",
					"available), command can also be specified directly using syntax 'cmd %s'",
					"where '%s' will be replaced with directory path, e.g. 'lf %s'",
				),
			)
		} else if strings.HasPrefix(action.Use, "info") {
			action.Flags().BoolP("json", "j", false, "output app info in JSON")
		}
		subcmd_app.AddCommand(action)
	}

	db_actions := [3]*cobra.Command{
		&cobra.Command{
			Use:     "check",
			Aliases: []string{"c"},
			Short:   "compare database versus installed apps, asks for update if found any changes",
			Args:    cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				if !made_a_new_db {
					db_action(cmd, args)
				}
			},
		},
		&cobra.Command{
			Use:     "last-update-time",
			Aliases: []string{"t", "timestamp"},
			Short:   "show last DB update time",
			Args:    cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				db_action(cmd, args)
			},
		},
		&cobra.Command{
			Use:     "update",
			Aliases: []string{"u"},
			Short:   "force update of database (will remove all existing data)",
			Args:    cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				if !made_a_new_db {
					db_action(cmd, args)
				}
			},
		},
	}
	for _, action := range db_actions {
		subcmd_db.AddCommand(action)
	}
}

func main() {
	if err := rootcmd.Execute(); err != nil {
		//fmt.Println(err)
		os.Exit(60)
	}
	os.Exit(0)
}
