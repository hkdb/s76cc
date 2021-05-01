#!/bin/bash

echo ""
echo "Checking to make sure ~/.local exists..."
echo ""

# Check for ~/.local
if [ ! -d "$HOME/.local" ]; then
   echo "There doesn't seem to be a .local directory in $HOME. This won't work. Please submit an issue to https://github.com/hkdb/s76cc/issues for more help... EXITING!"
   echo ""
   exit
else
   echo "Looks good!"
   echo ""
fi

# Copy files to location
echo "Installing files..."
echo ""
cp s76cc $HOME/.local/bin/
mkdir -p $HOME/.local/share/icons/3dfosi
cp s76cc.png $HOME/.local/share/icons/3dfosi/
cp s76cc.desktop $HOME/.local/share/applications/
echo "DONE!"
echo ""
