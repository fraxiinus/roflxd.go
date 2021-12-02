# roflxd.go - ROFL eXtract Data (Golang vers.)

[Go back to roflxd overview](https://github.com/fraxiinus/roflxd)

## About

**This is an incomplete project.** The original goal of this project was to convert [GLR replays](https://github.com/1lann/lol-replay) to a form that was playable by the League of Legends client. That goal ended up being unachievable so the project was shelved.

The ROFL parsing ability of the project is still useful, so this project is now a part of roflxd.

See original discussions on [ReplayBook](https://github.com/fraxiinus/ReplayBook/discussions/75).

This is literally my first ever golang project. I hope it is at the very least, readable. Best of luck.

## What is working

* Loading ROFL files to memory
* Saving data in memory to ROFL file
* Loading GLR files to memory

## How to use

Compile and run using go:

>go run .

Example 1: Reads ROFL file, verifies, and prints verbose logs

>go run . -input EUN1.rofl -v -mode verify

Example 2: Reads a ROFL file, and outputs all data to JSON file

>go run . -input NA1.rofl -mode json -output "dump.json"
