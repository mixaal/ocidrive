export OCI_DRIVE_COMPARTMENT_ID="ocid1.compartment.oc1..<your bucket comparmtment ocid>" 
export OCI_DRIVE_BUCKET_NAME="<bucket name>" # example: images
export OCI_DRIVE_LOCAL_FS="<localfs storage>" # example /home/bob/ObjectStore
export OCI_DRIVE_ID="<id>"

# This doesn't work on windows
OS=$(uname -s | tr A-Z a-z)
./ocidrive-${OS}-amd64
