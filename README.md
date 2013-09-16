# hankie

[Hankie ist ein Dockarbeiter](http://www.youtube.com/watch?v=K5f87QQjcbc)


## Bugs and design issues

This is very early work in progress / prototyping and has some serious issues. Some rather easy to fix, some not.

 - This requires https://github.com/dotcloud/docker/pull/1877
 - If a docker instance isn't refered by any job file, it's ignored
 - We're abusing the host/domainname to identify instances to avoid tracking the docker assigned ID (Proper solution blocked by dotcloud/docker#1)
 - Cleaning up instances is missing
 - Running several instances of the same job/env/product combination isn't supported
 - Update to existing jobs isn't possible (if job/env/product is running, it doesn't care whether other options changed)
 - Image download missing
 - Cleanup missing (killed instances are left on system)
 - Documentation and tests missing
