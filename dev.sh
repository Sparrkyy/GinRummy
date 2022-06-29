#!/bin/bash
tmux kill-session

session="system-startup"

tmux new-session -d -s $session

window=0
tmux rename-window -t $session:$window 'frontend'
tmux send-keys -t $session:$window 'cd frontend && npm run dev' C-m
tmux split-window 'cd backend && air'
tmux attach-session -t $session
