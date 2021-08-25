# Build the CLI app
make build

# Migrate the database
./sec migrate up

# Testing enabling downloads
./sec de 2015/10
./sec de 2021

# Listing worklist
./sec dlist

# Download index files
./sec dow index  

# Display space needed to download data files
./sec dest

# Display space needed to download ZIP data files
./sec destz