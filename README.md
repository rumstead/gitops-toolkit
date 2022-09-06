# argo-cd-toolkit
Helpful manifests, scripts, and tools for Argo CD

## Hack
Scripts under `hack/` need to be run from the `hack/` directory
```shell
cd hack/
./start.sh
```
### sed
Sed isn't the most portable binary. The scripts assume GNU sed. 

#### Mac users
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