# OCI Drive

... use your storage with Oracle Object Store

[![Build Status](https://travis-ci.com/mixaal/ocidrive.svg?branch=master)](https://travis-ci.com/mixaal/ocidrive)


## Quick Start

Make sure you have the Object Storage, bucket and you know the compartment id where the storage is configured and also have access there.

Create your OCI configuration, for more details refer to:

https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm

```
mkdir -p "$HOME/.oci"
cat <<EOF > "$HOME/.oci/config"
[DEFAULT]
user=ocid1.user.oc1..XXXX<Your USER ID>XXXXXX
fingerprint=YY:OO:UU:RR: :FF:II:NN:GG:EE:RR:PP:RR:II:NN:TT
tenancy=ocid1.tenancy.oc1..XXXXX<USER TENANCY>XXXXX
region=<REGION> # Example: us-phoenix-1
key_file=~/.oci/oci_api_key.pem
EOF
```

Copy your KEY into `$HOME/.oci/oci_api_key.pem`.

Export variables:
```
export OCI_DRIVE_COMPARTMENT_ID="ocid1.compartment.oc1..<your bucket comparmtment ocid>" 
export OCI_DRIVE_BUCKET_NAME="<bucket name>" # example: images
export OCI_DRIVE_LOCAL_FS="<localfs storage>" # example /home/bob/ObjectStore
export OCI_DRIVE_ID="<id>", #example marynek, needed, but comes useful when multiple drives ran on the same machine
```

Execute for your architecture:

```
./ocidrive-linux-amd64
```


##  Windows Instructions

1. Make the variables above as user environment variables: Settings > System > Advanced > Envinroment, Add user variable
2. Use full paths inc config, e.g. replace `~/.oci/config` with `C:\Users\micha\.oci\config`
3. Make shortcut from the `ocidrive-windows-amd64.exe`, Winkey+R, enter `shell:startup` and copy the shortcut into the window


## Binaries

Under `build` directory you can find the latest binaries.

## Issues

* when changing the local fs directory, please either remove the `$HOME/.oci/id/*` files or move the whole contents of the directory, otherwise the remote directory will be wiped by the _empty_ content of the new directory
* for sync only size is compared, the file is synced when the size differs
* ~~empty directories are removed from the local file system immediately when found~~
* diffs from last and current snapshots might not work properly in case they are deleted
* when context timeout deadline is exceeded the file is not upload/downloaded to/from storage
