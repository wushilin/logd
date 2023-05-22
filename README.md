# logd
Log swallower, rotater and more

# Getting binary
```sh
$ go install github.com/wushilin/logd@latest
```


# Using it
```sh
$ run-my-program | logd -out app.log -size 100M -keep 20
```

Receive logs from run-my-program, write to app.log and rotate every 100 MiB (pre-write checked)

# Adding dates to logs
If your log does not print date time, you can use `-dated` to add a timestamp for eachline for you automatically
```sh
$ run-my-program | logd -out app.log -size 100M -keep 20 -dated
```
