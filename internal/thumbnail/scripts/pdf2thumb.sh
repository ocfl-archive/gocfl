#!/bin/bash

# Default values
BACKGROUND="none"
WIDTH=256
HEIGHT=256

# Function to display usage
usage() {
    echo "Usage: $0 -Source <path> -Destination <path> [-Background <color>] [-Width <int>] [-Height <int>]"
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -Source)
            SOURCE="$2"
            shift 2
            ;;
        -Destination)
            DESTINATION="$2"
            shift 2
            ;;
        -Background)
            BACKGROUND="$2"
            shift 2
            ;;
        -Width)
            WIDTH="$2"
            shift 2
            ;;
        -Height)
            HEIGHT="$2"
            shift 2
            ;;
        *)
            usage
            ;;
    esac
done

# Check mandatory parameters
if [[ -z "$SOURCE" || -z "$DESTINATION" ]]; then
    usage
fi

# Temporary file for the initial PNG extraction
TEMP_PNG="${DESTINATION}.tmp.png"

# Extract first page using Ghostscript
%%GHOSTSCRIPT%% %%GHOSTSCRIPT_PARAMS%% -dNOPAUSE -dBATCH -sDEVICE=png16m -dFirstPage=1 -dLastPage=1 -sOutputFile="$TEMP_PNG" "$SOURCE"

# Resize and format using ImageMagick
# In Ubuntu, 'convert' is part of ImageMagick 6, 'magick' is part of ImageMagick 7.
# Usually 'convert' is available.
%%CONVERT%% %%CONVERT_PARAMS%% "$TEMP_PNG" -resize "${WIDTH}x${HEIGHT}" -background "$BACKGROUND" -gravity Center -extent "${WIDTH}x${HEIGHT}" "$DESTINATION"

# Remove temporary file
rm -f "$TEMP_PNG"
