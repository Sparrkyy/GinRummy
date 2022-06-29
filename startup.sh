#!/bin/bash
tmux kill-session

session="system-startup"

tmux new-session -d -s $session

window=0
tmux rename-window -t $session:$window 'frontend'
tmux send-keys -t $session:$window 'cd frontend && npm run build && npm run start' C-m
tmux split-window 'cd backend && go run .'
tmux attach-session -t $session
