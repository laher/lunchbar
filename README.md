# Lunchbar

**VERY EARLY POC - NOT READY YET**

Lunchbar is an [xbar](https://github.com/matryer/xbar) clone which is targeted for cross-platform use on Mac, Linux and Windows.

The main point of this project is to be able to easily share plugins and tools between team-mates, even if they're not using the same Operating System.

Lunchbar uses [lunchbox](https://github.com/laher/lunchbox) to provide batteries-included support for cross-platform plugins

## Comparison

What are the similarities and differences between xbar and lunchbar?

`lunchbar` is in very early stages, but this should cover the conceptual differences & similarities...

### Commonality

 * they're both written in go.
 * The plugin design and format is the same. I plan to add some metadata tags, but that's all (for now). Anything beyond that I'll try to upstream to xbar.
   * lunchbar depends on xbar packages, including the emoji support. So, the look and feel is very similar. On Mac at least.

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
   * Some features are hard to implement in a cross-platform way.
     * icons only for some platforms. The first item shows what would normally be in the list.
     * IPC - lunchbar uses getlantern/systray instead of wails. This is already cross-platform, but it's different... only one systray icon per process. So, this tool launches one process per icon.
 * Little differences:
   * The 'lunchbar menu' is at the top

### Hopes and dreams

 * Support for easily building cross-platform plugins in bash/elvish/go/etc
 * OS tagging for plugins.
 * Some mechanism to discover and install plugins. Possibly similar to the xbar wasy. Not sure.
 * Some form of automated ENV management (<plugin>.env files) with some encryption support.


## Installation

For now, it's not pre-packaged. Please install lunchbar using the Go compiler.

 1. Check getlantern/systray's docs for installing pre-requisites.
 2. Install lunchbar

    go install github.com/laher/lunchbar@latest

 3. Run lunchbar

    lunchbar

 4. Run lunchbar with a single plugin (plugins should be any executable located in $HOME/.config/lunchbar/plugins/)

    lunchbar myplugin.sh

## Progress

### M1

Focus of M1 is to get something usable working across OSes ...

 - [x] parse and handle plugin output
   - [x] href, shell, terminal[mac]
 - [x] 2 modes - supervisor and runner
   - [x] IPC
 - [x] Quit
 - [x] Refreshing
   - [x] Refresh button
   - [x] Refresh on timer
   - [x] Refresh all (IPC)
 - [x] Cross platform work
   - [x] Mac support
   - [x] Support icon instead of text. Basic.
   - [x] Verify
      - [x] Test on Linux
      - [x] Test on Windows
 - [ ] Error handling
   - [ ] Show a sample when no plugins found
   - [ ] React to killed process
 - [ ] Testing 
   - [ ] Happy Path Tests
   - [ ] Error path tests

### M2

M2 is to add some convenience features ahead of plugin management

 - [ ] Build automation (go-releaser cross compilation?)!
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
