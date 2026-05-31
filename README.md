# B-shell

a custom POSIX-compliant shell inspired by bash.


For now:
```zsh
go run main.go
```


it's valid to call me a hipster for using exec.Command() instead of doing the bread and butter of C's 
```c 
fork() 
execlp() 
waitpid()
``` 
and looking up the binary in the $PATH



ls | cat | echo | cmd
