#!/usr/bin/env bash

SESSION=demo

tmux -2 new-session -d -s $SESSION
tmux split-window -h
tmux select-pane -t 1
tmux resize-pane -R 30
tmux send-keys "sleep 170 && ./watcher/monkey.py --namespace=ccp --regexp='mariadb-percona-*' --period=25" C-m
tmux select-pane -t 0
tmux send-keys "./demo_19.09.sh" C-m

tmux -2 attach-session -t $SESSION
