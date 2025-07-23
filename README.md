# prm - cli audio plugin remover for mac/linux #

## Installation ##
Run the following command to install.

```
curl -sS https://raw.githubusercontent.com/asherqnovick/prm/main/install/install.sh | bash
```

then you can run
```
prm
```

To uninstall run
```
curl -sS https://raw.githubusercontent.com/asherqnovick/prm/main/install/uninstall.sh | bash
```

## Usage ##
Finds ```.aaxplugin``` ```.component``` ```.clap``` ```.vst``` ```.vst3``` and ```.driver``` files for the given paths

Search for plugins, filter by format, sort by size.

Delete plugins. Plugins are moved to the trash and can be manually restored.

Pipe output paths to other programs

## Paths ##
New paths can be added to ```$HOME/.config/prm/paths.txt```, each path on its own line.

This file will automatically be created if necessary and populated with known paths.

The flag ```-paths``` can be used to open this file.

If you want to reset to defaults, delete the file.

## Flags ##
The following flags can be set, followed by a case insensitive search term

Flags with full words will override all other options

  **-c** count results

  **-p** show paths

  **-s** sort results by size (automatically displays paths)

  **-f=string** narrow results by format (aax, au, clap, vst, vst3, driver)

  **-delete** move results to trash (will ask for confirmation)

  **-open** open all plugin paths

  **-paths** open paths.txt file


## Examples ##


```prm comp``` - display plugins containg 'comp'

```prm -p fab``` - outputs paths of all plugins containing 'fab'

```prm -c -f=clap chorus``` - count and output all clap plugins containing 'chorus'

```prm -f=vst -d Pro-Q 2``` - finds and deletes VST plugins that match the term Pro-Q 2

```prm -f=aax -delete``` - deletes all AAX plugins

```prm -p | less``` - output all paths to less for easy scrolling

```prm -s | sizes.txt``` - output all paths and their sizes (sorted) to sizes.txt
