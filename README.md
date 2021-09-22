# Crossbar

VERY EARLY POC - NOT READY YET

Crossbar is an [xbar](https://github.com/matryer/xbar) clone which is targeted for Mac, Linux and Windows.

## Clone talk

What are the similarities and differences between xbar and crossbar

Commonality:
 * they're both written in go
 * crossbar depends on xbar packages. So, the look and feel is very similar. On Mac at least.
 * The plugin design and format is the same. I plan to add some metadata tags, but that's all (for now).

Key differences from xbar:
 * Cross-platform from the outset.
  * Use [getlantern/systray](https://github.com/getlantern/systray) instead of wails.
    * xbar's UI library, Wails, doesn't currently support Windows or Linux. It could take a while before it gets implemented, and even then there are some challenges....
  * There are peculiar challenges when developing a cross-platform system tray.
    * OS differences - e.g. Windows doesn't show text.
  * Each plugin runs in its own process (this is forced by a limitation of systray). IPC communications between main app and plugin app.
 * Reduced feature set.
   * No plugin browser (for the time being)
   * Some features 


# Setup

For now, it's not pre-packaged. Please install crossbar using a Go compiler.

 1. Check getlantern/systray's docs for installation instructions.
 2. Install

    go install github.com/laher/crossbar@latest
