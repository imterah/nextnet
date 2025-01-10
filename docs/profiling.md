# Profiling
To profile any backend code based on `backendutil`, follow these steps:
1. Rebuild the backend with the `debug` flag: `cd $BACKEND_HERE; GOOS=linux go build -tags debug .; cd ..`
2. Copy the binary to the target machine (if applicable), and stop the API server.
3. If you want to profile the CPU utilization, write `cpu` to the file `/tmp/hermes.backendlauncher.profilebackends`: `echo -n "cpu" > /tmp/hermes.backendlauncher.profilebackends`. Else, replace `cpu` with `mem`.
4. Start the API server, with development mode and debug logging enabled.
