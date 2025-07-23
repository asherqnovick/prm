package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
)

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
)

var HOME_PATH = os.Getenv("HOME")
var CONFIG_PATH = filepath.Join(HOME_PATH, ".config", "prm")
var PATHS_PATH = filepath.Join(CONFIG_PATH, "paths.txt")

var PLUGIN_PATHS = map[string]bool{}
var MAC_PATHS = map[string]bool{
	"/Library/Audio/Plug-Ins":                          true,
	"/Library/Application Support/Avid/Audio/Plug-Ins": true,
}
var LINUX_PATHS = map[string]bool{
	"~/.clap/":             true,
	"/usr/lib/clap/":       true,
	"/usr/local/lib/clap/": true,
	"~/.vst/":              true,
	"/usr/lib/vst/":        true,
	"/usr/local/lib/vst/":  true,
	"~/.vst3/":             true,
	"/usr/lib/vst3/":       true,
	"/usr/local/lib/vst3/": true,
}
var WINDOWS_PATHS = map[string]bool{
	`C:\Program Files\Common Files\Avid\Audio\Plug-Ins`:       true,
	`C:\Program Files (x86)\Common Files\Avid\Audio\Plug-Ins`: true,
	`C:\Program Files\Vstplugins\`:                            true,
	`C:\Program Files\Steinberg\VSTPlugins\`:                  true,
	`C:\Program Files (x86)\Vstplugins\`:                      true,
	`C:\Program Files (x86)\Steinberg\VSTPlugins\`:            true,
	`C:\Program Files\Common Files\VST3\`:                     true,
	`C:\Program Files (x86)\Common Files\VST3\`:               true,
	`C:\Program Files\Common Files\CLAP\`:                     true,
	`C:\Program Files (x86)\Common Files\CLAP\`:               true,
}

type Plugins = map[string]*Plugin
type PluginList = []*Plugin
type PluginGroups = map[string]PluginList
type Plugin struct {
	Name  string
	Paths []string
}

func main() {
	count := flag.Bool("c", false, "count results")
	delete := flag.Bool("delete", false, "move plugins to trash (will ask for confirmation)")
	format := flag.String("f", "", "narrow results by format (aax, au, clap, vst, vst3, driver)")
	open := flag.Bool("open", false, "open all plugin paths")
	path := flag.Bool("p", false, "print paths of results")
	paths := flag.Bool("paths", false, "open paths.txt file")
	sort := flag.Bool("s", false, "sort results by size (automatically displays paths)")
	flag.Parse()
	args := flag.Args()

	loadConfig()

	if *open {
		openFoldersInExplorer()
		return
	}
	if *paths {
		openPathsFile()
		return
	}

	var pp Plugins = scanPaths()
	pp = searchPlugins(strings.Join(args, " "), pp)

	if *format != "" {
		str := *format
		pp = searchFormat(str, pp)
	}
	if !*delete {
		if *sort {
			printPathsBySize(pp)
		} else if *path {
			printPaths(pp)
		} else {
			printPluginsByName(pp)
		}
		if *count == true {
			fmt.Println(countPlugins(pp))
		}
	} else if *delete {
		fmt.Print(Red)
		for _, p := range pp {
			for _, path := range p.Paths {
				fmt.Println(path)
			}
		}
		fmt.Print(Reset)

		var confirm string
		fmt.Print(Yellow + "Do you REALLY want to delete all of these paths? (type 'delete' to confirm): " + Reset)
		fmt.Scan(&confirm)
		switch confirm {
		default:
			fmt.Print(Green)
			fmt.Println("Aborted. Nothing deleted.")
			fmt.Print(Reset)
			return
		case "delete":
			for _, p := range pp {
				for _, path := range p.Paths {
					trashPath(path)
					fmt.Println(path, "- "+Red+"DELETED"+Reset)
				}
			}
		}
	}
}
func loadConfig() {
	_, err := os.Stat(CONFIG_PATH)
	if os.IsNotExist(err) {
		os.Mkdir(CONFIG_PATH, 0755)
	}
	_, err = os.Stat(PATHS_PATH)
	if os.IsNotExist(err) {
		file, _ := os.Create(PATHS_PATH)
		switch runtime.GOOS {
		case "darwin":
			for path := range MAC_PATHS {
				fmt.Fprintf(file, "%s\n", path)
			}
		case "linux":
			for path := range LINUX_PATHS {
				fmt.Fprintf(file, "%s\n", path)
			}
		}
	}
	file, err := os.Open(PATHS_PATH)
	if err != nil {
		fmt.Println("error creating paths file, unable to continue")
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = expandPath(line)
		_, exists := PLUGIN_PATHS[line]
		if !exists {
			PLUGIN_PATHS[line] = true
		}
	}
	file.Close()
}
func scanPath(path string) (pp Plugins) {
	pp = make(Plugins)
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	for _, entry := range dirEntries {
		fp := filepath.Join(path, entry.Name())
		entryName := entry.Name()
		ext := filepath.Ext(entryName)
		entryBasename := strings.TrimSuffix(entryName, ext)

		switch ext {
		case "":
			subPlugins := scanPath(fp)

			for name, subPlugin := range subPlugins {
				if existingPlugin, exists := pp[name]; exists {
					existingPlugin.Paths = append(existingPlugin.Paths, subPlugin.Paths...)
				} else {
					pp[name] = subPlugin
				}
			}

		case ".aaxplugin", ".clap", ".component", ".vst", ".vst3", ".driver":
			if p, exists := pp[entryBasename]; exists {
				p.Paths = append(p.Paths, fp)
			} else {
				pp[entryBasename] = &Plugin{
					Name:  entryBasename,
					Paths: []string{fp},
				}
			}
		}
	}

	return
}
func scanPaths() (pp Plugins) {
	pp = make(Plugins)
	for p := range PLUGIN_PATHS {
		subPlugins := scanPath(p)

		for name, subPlugin := range subPlugins {
			if existingPlugin, exists := pp[name]; exists {
				existingPlugin.Paths = append(existingPlugin.Paths, subPlugin.Paths...)
			} else {
				pp[name] = subPlugin
			}
		}
	}
	return
}
func searchPlugins(search string, pp Plugins) (result Plugins) {
	result = make(Plugins)

	for pName, plugin := range pp {
		if strings.Contains(strings.ToLower(pName), strings.ToLower(search)) {
			result[pName] = plugin
		}
	}
	return
}
func searchFormat(format string, pp Plugins) (result Plugins) {
	result = make(Plugins)
	for pName, plugin := range pp {
		for _, f := range plugin.Formats() {
			if strings.ToLower(f) == strings.ToLower(format) {
				plugin.NarrowPaths(f)
				result[pName] = plugin
			}
		}
	}
	return
}
func (p *Plugin) NarrowPaths(format string) {
	narrowed := []string{}
	for _, p := range p.Paths {
		if strings.ToLower(format) == "au" {
			if filepath.Ext(p) == ".component" {
				narrowed = append(narrowed, p)
			}
		} else if strings.ToLower(format) == "aax" {
			if filepath.Ext(p) == ".aaxplugin" {
				narrowed = append(narrowed, p)
			}
		} else if filepath.Ext(p) == "."+strings.ToLower(format) {
			narrowed = append(narrowed, p)
		}
	}
	p.Paths = narrowed
}
func printPluginsByName(pp Plugins) {
	names := make([]string, 0, len(pp))
	for pName := range pp {
		names = append(names, pName)
	}
	slices.SortFunc(names, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	for _, pName := range names {
		if p, exists := pp[pName]; exists {
			fmt.Println(p.Name, p.Formats())

		}
	}
}
func printPaths(pp Plugins) {
	for _, p := range pp {
		for _, path := range p.Paths {
			fmt.Println(path)
		}
	}
}
func printPathsBySize(pp Plugins) {

	type pathSize struct {
		path string
		size int64
	}
	var pathSizes []pathSize

	for _, p := range pp {
		for _, path := range p.Paths {
			size, err := getDirSize(path)
			if err != nil {
				fmt.Printf("Error getting size for %s: %v\n", path, err)
				continue
			}
			pathSizes = append(pathSizes, pathSize{path, size})
		}
	}
	sort.Slice(pathSizes, func(i, j int) bool {
		return pathSizes[i].size > pathSizes[j].size
	})

	for _, ps := range pathSizes {
		sizeMB := float64(ps.size) / (1024 * 1024)
		fmt.Printf("%s %.2fMB\n", ps.path, sizeMB)
	}
}
func countPlugins(pp Plugins) int {
	count := 0
	for _, p := range pp {
		for range p.Paths {
			count += 1
		}
	}
	return count
}
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
func trashPath(path string) {
	switch runtime.GOOS {
	case "darwin":
		move := exec.Command("sudo", "mv", path, filepath.Join(HOME_PATH, ".Trash", filepath.Base(path)))
		err := move.Run()
		if err != nil {
			fmt.Println(err)
		}
	case "linux":
		move := exec.Command("sudo", "mv", path, filepath.Join(HOME_PATH, ".local/share/Trash/files", filepath.Base(path)))
		err := move.Run()
		if err != nil {
			fmt.Println(err)
		}
	}
}
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(HOME_PATH, path[2:])
	}
	return path
}
func openFoldersInExplorer() {
	for path, _ := range PLUGIN_PATHS {
		cmd := exec.Command("open", path)
		cmd.Start()
	}
}
func openPathsFile() {
	cmd := exec.Command("open", PATHS_PATH)
	cmd.Start()
}
func (p Plugin) Formats() []string {
	ff := []string{}
	for _, path := range p.Paths {
		switch filepath.Ext(path) {
		case ".aaxplugin":
			ff = append(ff, "AAX")
		case ".clap":
			ff = append(ff, "CLAP")
		case ".component":
			ff = append(ff, "AU")
		case ".vst":
			ff = append(ff, "VST")
		case ".vst3":
			ff = append(ff, "VST3")
		case ".driver":
			ff = append(ff, "Driver")
		}
	}
	return ff
}
