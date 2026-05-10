+++
title = "exit"
chapter = false
weight = 106
hidden = false
+++

## Summary
Perform self deletion and exit the process.

  
- Needs Admin: False  
- Version: 1  
- Author: @brmk  

### Arguments

## Usage

```
exit
```


## Detailed Summary

Performs self deletion and exits the process:
* **windows** - self deletion is performed by renaming the file to an Alternate Data Stream (ADS) (default: ":pikka") and deleting it.
* **unix/macos** - self deletion is more crunchy, it forks a process to delete the file and exits the process.