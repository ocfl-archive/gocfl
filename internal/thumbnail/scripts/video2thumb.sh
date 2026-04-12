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

# Temporary file for the initial frame extraction
TEMP_PNG="${DESTINATION}.tmp.png"

# Extract frame at 00:00:35 using ffmpeg
%%FFMPEG%% %%FFMPEG_PARAMS%% -ss 00:00:35 -i "$SOURCE" -frames:v 1 "$TEMP_PNG"

# Resize and format using ImageMagick
%%CONVERT%% %%CONVERT_PARAMS%% "$TEMP_PNG" -resize "${WIDTH}x${HEIGHT}" -background "$BACKGROUND" -gravity Center -extent "${WIDTH}x${HEIGHT}" "$DESTINATION"

# Remove temporary file
rm -f "$TEMP_PNG"
