#!/usr/bin/env python3

import os, time, subprocess, sys

def runCmd(cmd):
    p = subprocess.Popen(cmd, shell = True, stdout = subprocess.PIPE, stderr = subprocess.PIPE)
    stdout = p.communicate()[0].decode('utf-8').strip()
    return stdout

# Get last tag.
def lastTag():
    return runCmd('git describe --abbrev=0 --tags')

# Get current branch name.
def branch():
    return runCmd('git rev-parse --abbrev-ref HEAD')

# Get last git commit id.
def lastCommitId():
    return runCmd('git log --pretty=format:"%h" -1')

# Assemble build command.
def buildCmd():
    buildFlag = []

    version = lastTag()
    if version != "":
        buildFlag.append("-X main._version_ '{}'".format(version))

    branchName = branch()
    if branchName != "":
        buildFlag.append("-X main._branch_ '{}'".format(branchName))

    commitId = lastCommitId()
    if commitId != "":
        buildFlag.append("-X main._commitId_ '{}'".format(commitId))

    # current time
    buildFlag.append("-X main._buildTime_ '{}'".format(time.strftime("%Y-%m-%d %H:%M %z")))

    return 'go build -ldflags "{}"'.format(" ".join(buildFlag))

cmd = buildCmd()

if len(sys.argv) > 1 and sys.argv[1] == "-showcmd":
    print(cmd)
elif subprocess.call(cmd, shell = True) == 0:
    print("build finished.")

