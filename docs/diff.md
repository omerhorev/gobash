This file will specify all the differences between IEEE1003.1 shell specification and this implementation.

# Error Handling
TODO: Compare

# Builtin
- Builtins ignore io redirects. For example, `cd /tmp/folder > /tmp/file` will not create the file /tmp/file at all.
    - Except the builtin `exec`

# Default Commands
Some of the functionality of GNU coreutils is implemented as default commands.
Check out the code documentation of the commands in the cmd package