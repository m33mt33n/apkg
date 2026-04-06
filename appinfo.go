package apkg

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/m33mt33n/apkg/mparser"
)

type Package string
type ApkInfo struct {
	Package       Package `json:"package"`
	Name          string  `json:"name"`
	Version       string  `json:"version"`
	Version_code  string  `json:"version_code"`
	Main_activity string  `json:"main_activity"`
	Minimum_SDK   uint8   `json:"minimum_sdk"`
	Target_SDK    uint8   `json:"target_sdk"`
}
type InstallationInfo struct {
	Apk_path        string
	Data_dir        string
	Library_path    string
	Resource_path   string
	UID             uint
	First_installed time.Time
	Last_update     time.Time
}
type App struct {
	Package         Package   `json:"package"`
	Name            string    `json:"name"`
	Version         string    `json:"version"`
	Version_code    string    `json:"version_code"`
	Main_activity   string    `json:"main_activity"`
	Minimum_SDK     uint8     `json:"minimum_sdk"`
	Target_SDK      uint8     `json:"target_sdk"`
	UID             uint      `json:"uid"`
	First_installed time.Time `json:"first_installed"`
	Last_update     time.Time `json:"last_update"`
	Apk_path        string    `json:"apk_path"`
	Data_dir        string    `json:"data_dir"`
	Library_path    string    `json:"library_path"`
	Resource_path   string    `json:"resource_path"`
}

var (
	RE_Install_info = [7]*regexp.Regexp{
		regexp.MustCompile("codePath=(.+)"),
		regexp.MustCompile("dataDir=(.+)"),
		regexp.MustCompile("nativeLibraryPath=(.+)"),
		regexp.MustCompile("resourcePath=(.+)"),
		regexp.MustCompile("userId=(\\d+) "),
		regexp.MustCompile("firstInstallTime=(.+)"),
		regexp.MustCompile("lastUpdateTime=(.+)"),
	}
	RE_Appinfo = [7]*regexp.Regexp{
		regexp.MustCompile("package: name='(.*?)'"),
		regexp.MustCompile("application-label:'(.*?)'"),
		regexp.MustCompile("versionName='(.*?)'"),
		regexp.MustCompile("versionCode='(.*?)'"),
		regexp.MustCompile("launchable-activity: name='(.*?)'"),
		regexp.MustCompile("sdkVersion:'(\\d*?)'"),
		regexp.MustCompile("targetSdkVersion:'(\\d*?)'"),
	}
	Aapt_path = "aapt"
)

var appinfo string = `
Package Name     %s
App Name         %s
Version          %s
Version Code     %s
Main Activity    %s
Minimum SDK      %s
Target SDK       %s
`

func get_mactivity_override(app App, debug bool) App {
	// get overrided main activities
	os.MkdirAll(Data_dir, 0750)
	overrides_file := filepath.Join(Data_dir, "mactivity_override.json")
	var (
		data_bytes []byte
		overrides  map[string]string
		err        error
	)
	if File_exists(overrides_file) {
		if data_bytes, err = os.ReadFile(overrides_file); err == nil {
			err = json.Unmarshal(data_bytes, &overrides)
		}
	} else {
		var file *os.File
		overrides = map[string]string{
			"com.dummy.package":  "dummy.main.activity",
		}
		if file, err = os.Create(overrides_file); err == nil {
			defer file.Close()
			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(overrides)
		}
	}
	if err == nil {
		for pkg, act := range overrides {
			if Package(pkg) == app.Package {
				app.Main_activity = act
			}
		}
	} else if debug {
		fmt.Println(err)
	}
	return app
}

func (pkg Package) Kill(safe_kill bool, debug bool) uint8 {
	var cmd string
	if safe_kill {
		cmd = "am kill " + string(pkg)
	} else {
		cmd = "pkill -KILL " + string(pkg)
	}
	output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if debug {
		fmt.Println(string(output))
		if err != nil {
			fmt.Println(err)
		}
	}
	if err != nil {
		return 129
	}
	return 0
}

func (pkg Package) Force_stop(debug bool) uint8 {
	output, err := exec.Command("sh", "-c", "am force-stop "+string(pkg)).CombinedOutput()
	if debug {
		fmt.Println(string(output))
		if err != nil {
			fmt.Println(err)
		}
	}
	if err != nil {
		return 129
	}
	return 0
}

func (pkg Package) Change_state(state string, debug bool) uint8 {
	if state != "enable" && state != "disable" {
		return 1
	}
	output, err := exec.Command("sh", "-c", fmt.Sprintf("pm %s %s", state, string(pkg))).CombinedOutput()
	if debug {
		fmt.Println(string(output))
		if err != nil {
			fmt.Println(err)
		}
	}
	if err != nil {
		return 129
	}
	return 0
}

func (app App) String() string {
	Minimum_SDK := fmt.Sprintf("%d", app.Minimum_SDK)
	Target_SDK := fmt.Sprintf("%d", app.Target_SDK)
	if Minimum_SDK == "0" {
		Minimum_SDK = "NA"
	}
	if Target_SDK == "0" {
		Target_SDK = "NA"
	}
	app = get_mactivity_override(app, false)
	result := strings.Trim(
		fmt.Sprintf(
			appinfo,
			app.Package, app.Name, app.Version, app.Version_code,
			app.Main_activity, Minimum_SDK, Target_SDK,
		),
		"\n",
	)
	if app.UID != 0 {
		result += fmt.Sprintf("\nUID              %d", app.UID)
	}
	var time_zero time.Time
	if app.First_installed != time_zero {
		result += fmt.Sprintf(
			"\nFirst Installed  %s",
			app.First_installed.Format(time.DateTime),
		)
	}
	if app.Last_update != time_zero {
		result += fmt.Sprintf(
			"\nLast Update      %s",
			app.Last_update.Format(time.DateTime),
		)
	}
	return result
}

func (apkinfo ApkInfo) String() string {
	Minimum_SDK := fmt.Sprintf("%d", apkinfo.Minimum_SDK)
	Target_SDK := fmt.Sprintf("%d", apkinfo.Target_SDK)
	if Minimum_SDK == "0" {
		Minimum_SDK = "NA"
	}
	if Target_SDK == "0" {
		Target_SDK = "NA"
	}
	result := strings.Trim(
		fmt.Sprintf(
			appinfo,
			apkinfo.Package, apkinfo.Name, apkinfo.Version, apkinfo.Version_code,
			apkinfo.Main_activity, Minimum_SDK, Target_SDK,
		),
		"\n",
	)
	return result
}

func (app App) Start(debug bool) uint8 {
	app = get_mactivity_override(app, debug)
	output, err := exec.Command(
		"sh", "-c", fmt.Sprintf("am start '%s/%s'", string(app.Package), app.Main_activity),
	).CombinedOutput()
	if debug {
		fmt.Println(string(output))
		if err != nil {
			fmt.Println(err)
		}
	}
	if err != nil {
		return 129
	}
	return 0
}

func Get_installation_info(pkg Package) (InstallationInfo, uint8) {
	var inst_info InstallationInfo
	output, err := exec.Command("dumpsys", "package", string(pkg)).Output()
	if err != nil {
		return inst_info, 10
	}
	var sink [7]any
	var time_zero time.Time
	for idx, pattern := range RE_Install_info {
		result := pattern.FindStringSubmatch(string(output))
		if len(result) != 2 {
			if idx == 4 {
				sink[idx] = uint(0)
			} else if idx == 5 || idx == 6 {
				sink[idx] = time_zero
			} else {
				sink[idx] = "NA"
			}
			continue
		}
		if idx == 4 {
			var r int
			if r, err = strconv.Atoi(result[1]); err != nil {
				r = 0
			}
			sink[idx] = uint(r)
		} else if idx == 5 || idx == 6 {
			sink[idx], _ = time.ParseInLocation(time.DateTime, result[1], time.Local)
		} else {
			sink[idx] = result[1]
		}
	}
	inst_info.Apk_path = sink[0].(string)
	inst_info.Data_dir = sink[1].(string)
	inst_info.Library_path = sink[2].(string)
	inst_info.Resource_path = sink[3].(string)
	inst_info.UID = sink[4].(uint)
	inst_info.First_installed = sink[5].(time.Time)
	inst_info.Last_update = sink[6].(time.Time)
	//fmt.Println(inst_info)
	return inst_info, 0
}

func get_info(apk string) (ApkInfo, uint8) {
	var (
		apkinfo ApkInfo
		err error
	)
	output, err := exec.Command(Aapt_path, "dump", "badging", apk).Output()
	// m3: this error must be ignored as some (6/302) of apks are processed with errors by aapt
	// but command output still have usable data
	//if err != nil {
	//return apkinfo, 11
	//}
	var sink [7]any
	for idx, pattern := range RE_Appinfo {
		result := pattern.FindStringSubmatch(string(output))
		if len(result) != 2 || (len(result) == 2 && len(result[1]) == 0) {
			if idx == 5 || idx == 6 {
				sink[idx] = uint8(0)
			} else {
				sink[idx] = "NA"
			}
			continue
		}
		if idx == 5 || idx == 6 {
			var r int
			if r, err = strconv.Atoi(result[1]); err != nil {
				r = 0
			}
			sink[idx] = uint8(r)
		} else {
			sink[idx] = result[1]
		}
	}
	apkinfo.Package = Package(sink[0].(string))
	apkinfo.Name = sink[1].(string)
	apkinfo.Version = sink[2].(string)
	apkinfo.Version_code = sink[3].(string)
	apkinfo.Main_activity = sink[4].(string)
	if apkinfo.Main_activity == "NA" {
		activity_name, rcode := mparser.Get_main_activity(apk, string(apkinfo.Package))
		if rcode == 0 && len(activity_name) > 0 {
			apkinfo.Main_activity = activity_name
		}
	}
	apkinfo.Minimum_SDK = sink[5].(uint8)
	apkinfo.Target_SDK = sink[6].(uint8)
	return apkinfo, 0
}

func Get_info(pkg Package) (App, uint8) {
	var (
		app App
		apkinfo ApkInfo
	)
	inst_info, rcode := Get_installation_info(pkg)
	if rcode != 0 {
		return app, rcode
	}
	if inst_info.Apk_path == "NA" {
		return app, 12
	}
	apkinfo, rcode = get_info(inst_info.Apk_path)
	if rcode != 0 {
		return app, rcode
	}
	app.Package = apkinfo.Package
	app.Name = apkinfo.Name
	app.Version = apkinfo.Version
	app.Version_code = apkinfo.Version_code
	app.Main_activity = apkinfo.Main_activity
	app.Minimum_SDK = apkinfo.Minimum_SDK
	app.Target_SDK = apkinfo.Target_SDK
	app.Apk_path = inst_info.Apk_path
	app.Data_dir = inst_info.Data_dir
	app.Library_path = inst_info.Library_path
	app.Resource_path = inst_info.Resource_path
	app.UID = inst_info.UID
	app.First_installed = inst_info.First_installed
	app.Last_update = inst_info.Last_update
	return app, 0
}

func Get_info_from_apk(apk string) (ApkInfo, uint8) {
	var (
		apkinfo ApkInfo
		rcode   uint8
	)
	if !File_exists(apk) {
		return apkinfo, 13
	}
	// NOTE: main activity overrides currently not applied to `info from apk` feature.
	apkinfo, rcode = get_info(apk)
	if rcode != 0 {
		return apkinfo, rcode
	}
	return apkinfo, 0
}
