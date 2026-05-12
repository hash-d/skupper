
This playbook will verify that the Skupper binaries present on the system (or on the images) are good, in the sense that they do not dump core on the simplest invocations (generally the calls to show the help message and the command version).

No functionality is tested, and the output is not verified in any way.  This is the shallowest smoke test for the binaries, which only indicate they've been built properly.

There are tags on the playbook, which allows for either local binaries or images to be verified:

    ansible-playbook test.yaml --tags images -i arm-host,amd-host,z-host,p-host -vv

    ansible-playbook test.yaml --tags local_binaries -i localhost, -vv

The list of images and binaries to be verified are defined on the file `vars.yaml`, for both local binaries and images.

To access the images, the playbook expects that the image variables (such as `SKUPPER_ROUTER_IMAGE`) be defined locally.
