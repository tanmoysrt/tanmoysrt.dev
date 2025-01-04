# !/bin/bash

# Build the project
echo "Building the project..."

cd backend && go build -o blogger .

echo "Project built successfully!"
# Move the binary to current directory
cd ..
mv backend/blogger .