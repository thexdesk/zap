# zap
zap zap

# Usage

```sh
# Generate a 1gb file of random bytes
$ dd if=/dev/urandom of=src bs=1048576 count=1000

# Wait for zaps
$ zap --listen > dst
11:40PM INF listening for zaps addr=[::]:38197
```

```sh
# Zap "src" to its destination
$ zap src localhost:38197
11:41PM DBG zapped file bytes=1048576000
```
