#!/bin/bash

set -e

if [ -d "node_modules" ]; then
  echo "Dependencies already installed. Skipping installation."
  exit 0
fi

echo "Installing dependencies..."
npm install

echo "Ensuring react-router-dom is installed..."
npm install react-router-dom

echo "All dependencies installed."
