#!/bin/bash

FE='\033[0;31m' # Fatal error
NC='\033[0m' # No color
SU='\033[32m' # Success
NW='\033[33m' # Normal warning

if [ "$EUID" -ne 0 ]
	then echo -e "${FE}ERROR: ${NC}Run me with sudo!"
	exit
fi

if ([ -f watg ] && [ -f config.toml ])
	then echo -e "${SU}Files checked.${NC}"
else
	echo -e "${FE}ERROR: ${NC}File(s) missing or wrong type! Make sure the files ./watg and ./config.toml are in this folder."
	exit
fi

BG='\033[1;32m'   
BC='\033[1;36m'
BW='\033[1;37m'

echo -e "

 ${BG} _      __    ${BC}______     ${BW}__   _      __     ${BC}      ${BW}___       __ 
 ${BG}| | /| / /__ ${BC}/_  __/__ _${BW}/ /  (_)__  / /__  ${BC}      ${BW}/ _ )___  / /_
 ${BG}| |/ |/ / _ '/${BC}/ / / _ '/${BW} /__/ / _ \/  '_/ ${BC}      ${BW}/ _  / _ \/ __/
 ${BG}|__/|__/\_,_/${BC}/_/  \_, /${BW}____/_/_//_/_/\_\ ${BC}____  ${BW}/____/\___/\__/ 
 ${BG}            ${BC}     /___/${BW}                  ${BC}/___/ ${BW}                 

 ${BG}By MassiveBox, 2020. Released under GNU v3.0 - Release 0.1.0${NC}
"

while [ true ]
do
	echo "$(date +%H:%M:%S) - Starting bot…"
	./watg
	echo -e "$(date +%H:%M:%S) - ${NW}Bot has crashed!${NC} Restarting..."
done

