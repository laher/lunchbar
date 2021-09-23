# Crossbar

**VERY EARLY POC - NOT READY YET**

Crossbar is an [xbar](https://github.com/matryer/xbar) clone which is targeted for cross-platform use on Mac, Linux and Windows.

The main point of this project is to be able to share plugins and tools between team-mates, even if they're not using the same Operating System.

## Comparison

What are the similarities and differences between xbar and crossbar?

`crossbar` is in very early stages, but this should cover the conceptual differences & similarities...

### Commonality

 * they're both written in go
 * The plugin design and format is the same. I plan to add some metadata tags, but that's all (for now).
   * crossbar depends on xbar packages. So, the look and feel is very similar. On Mac at least.

### Key differences from xbar

 * Cross-platform from the outset.
    * Use [getlantern/systray](https://github.com/getlantern/systray) instead of wails.
      * xbar's UI library, Wails, doesn't currently support Windows or Linux. It could take a while before it gets implemented, and even then there are some challenges....
    * There are peculiar challenges when developing a cross-platform system tray.
      * OS differences - e.g. Windows doesn't show text.
    * Each plugin runs in its own process (this is forced by a limitation of systray). IPC communications between main app and plugin app.
    * Hope to write some support for building cross-platform plugins easily. e.g. a bundled `elvish` interpreter
 * Reduced feature set.
   * No plugin browser or other UI (for the time being)
   * Some features might be hard to implement in a cross-platform way. 

### Hopes and dreams

 * Some form of automated ENV management (<plugin>.env files) with some encryption support.
 * Some form of OS tagging for plugins.
 * Some support for easily building cross-platform plugins in bash/elvish/go/etc
 * Some mechanism to discover and install plugins. Possibly similar to the xbar wasy. Not sure.


## Installation

For now, it's not pre-packaged. Please install crossbar using the Go compiler.

 1. Check getlantern/systray's docs for installing pre-requisites.
 2. Install crossbar

    go install github.com/laher/crossbar@latest

 3. Run crossbar

    crossbar

 4. Run crossbar with a single plugin (plugins should be any executable located in $HOME/.config/crossbar/plugins/)

    crossbar myplugin.sh

## Progress

### M1

Focus of M1 is to get something usable working across OSes ...

 - [x] parse and handle plugin output
   - [x] href, shell, terminal[mac]
 - [x] 2 modes - supervisor and runner
   - [ ] IPC
 - [x] Quit
 - [ ] Refreshing
   - [ ] Refresh button
   - [ ] Refresh on timer
   - [ ] Refresh all (IPC)
 - [ ] Cross platform work
   - [x] Mac support
   - [ ] Support icon instead of text. Basic.
   - [ ] Verify
      - [ ] Test on Linux
      - [ ] Test on Windows
 - [ ] Error handling
   - [ ] Show a sample when no plugins found
   - [ ] React to killed process
 - [ ] Testing 
   - [ ] Happy Path Tests
   - [ ] Error path tests
 - [ ] Build automation?!

### M2

M2 is to add some convenience features ahead of plugin management

 - [ ] .env files
     - [ ] encrypted keys
 - [ ] Plugin-writing help 
   - [ ] Bundled-elvish support
   - [ ] Cross-platform convenience methods
   - [ ] Guides

### M3

M3 is all about plugin repo and managing compatibility with xbar plugins

 - [ ] Sample 'services'
 - [ ] Samples of some cross-platform [elvish] scripts
     - [ ] docker-compose
     - [ ] kube
     - [ ] weather
     - [ ] finance
