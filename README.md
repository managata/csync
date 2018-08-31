# csync

'csync' is a multi-node, multi-process rsync executer for copying large amount of files on clustered file system.

## Overview

If you have total 100TB files on a clustered file system and want to rsync them to another fs, you probably need to run multiple rsync processes on multiple hosts. 'Csync' will help you in such situation. For example,

    source directory: /fs1/src
    destination directory: /fs2/dst
    master host which csync runs on and which can read /fs1/src: csync_host
    source hosts which can read /fs1/src: src_host_a, src_host_b
    destination hosts which can write /fs2/dst: dst_host_c, dst_host_d
    source directory tree:
      /fs1/src/
      /fs1/src/a/
      /fs1/src/b/
      /fs1/src/b/c
      /fs1/src/b/d
    commandline: ./csync -p 2 -r src_host_a:dst_host_c -r src_host_b:dst_host_d -s /fs1/src -d /fs2/dst
    
    +----------+  +------------+                         +----------+ 
    |          |--| csync_host |--ssh--------+           |          |
    | /fs1/src |  |            |--ssh------+ |           | /fs2/dst |
    |          |  +------------+           | |           |          |
    |          |                           | |           |          |
    |          |  +------------+         +-+----------+  |          |
    |          |--| src_host_a |--rsync--| dst_host_c |--|          |
    |          |  |            |--rsync--|            |  |          |
    |          |  +------------+         +------------+  |          |
    |          |                             |           |          |
    |          |  +------------+         +---+--------+  |          |
    |          |--| src_host_b |--rsync--| dst_host_d |--|          |
    |          |  |            |--rsync--|            |  |          |
    +----------+  +------------+         +------------+  +----------+
    
    csync executes:
     1a. ssh dst_host_c rsync -dgloptADHX src_host_a:/fs1/src/ /fs2/dst/
     1b. ssh dst_host_c rsync -dgloptADHX src_host_a:/fs1/src/a/ /fs2/dst/a/
     1c. ssh dst_host_d rsync -dgloptADHX src_host_b:/fs1/src/b/ /fs2/dst/b/
     1d. ssh dst_host_d rsync -dgloptADHX src_host_b:/fs1/src/b/c/ /fs2/dst/b/c/
    
     (after 1a. completed)
     2a. ssh dst_host_c rsync -dgloptADHX src_host_a:/fs1/src/b/d/ /fs2/dst/b/d
    
As you can see above, each rsync will copy files in a directory but not sub-directories.

