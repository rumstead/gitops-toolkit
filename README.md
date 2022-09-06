# argo-cd-toolkit
Helpful manifests, scripts, and tools for Argo CD

## Setup multiple clusters connected to a central Argo CD
[start.sh](hack/multiple-clusters/start.sh) is the entrypoint to launching 4 k3d clusters, dev, qa, tst, and admin. 
Admin is the cluster which runs Argo CD and connects to all the target clusters (dev, qa, tst). 

### Configuration
**A lot** of things are hardcoded. Cluster names, Argo CD manifests (using HEAD), 


### Running the script
```shell
cd hack/multiple-clusters
./start.sh
```

#### Known issues
Sed isn't the most portable binary. The scripts assume GNU sed.
##### Mac users
Won't work :(
```shell
which sed       
# Bad
/usr/bin/sed
```
Will work
```shell
brew install gnu-sed
PATH="/usr/local/opt/gnu-sed/libexec/gnubin:$PATH" which sed
# Good
/usr/local/opt/gnu-sed/libexec/gnubin/sed
```

### Roadmap
1. Move the commands to Golang
2. Support more configuration and less hardcoded values