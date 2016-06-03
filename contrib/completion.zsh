#compdef jp jump

local profiles jump_commands

jump_commands=(
    -c:"create a new jump point"
    -m:"modify a jump point"
    -r:"remove a jump point"
)
_describe -t command "jump commands" jump_commands

profiles=($(ls -1 ~/.config/jump/))
compadd $profiles
