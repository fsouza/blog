+++
title = "Using Juju to orchestrate CentOS-based cloud services"
aliases = ["/2012/07/using-juju-to-orchestrate-centos-based.html"]

[taxonomies]
tags = ["centos", "juju", "python"]
+++

Earlier this week I had the opportunity to meet [Kyle
MacDonald](https://twitter.com/kylemacdonald/), head of [Ubuntu
Cloud](http://cloud.ubuntu.com/), during [FISL](http://fisl.org.br/13), and he
was surprised when we told him we are [using Juju with
CentOS](https://lists.ubuntu.com/archives/juju-dev/2012-June/000001.html) at
Globo.com. Then I decided to write this post explaining how we came up with a
[patched version of Juju](https://github.com/globocom/juju-centos-6) that
allows us to have CentOS clouds managed by Juju.<a name="more"></a>

For those who doesn't know Juju, it's a service orchestration tool, focused on
[devops](https://en.wikipedia.org/wiki/DevOps) "development method". It allows
you to deploy services on clouds, local machine and even bare metal machines
(using Canonical's MAAS).

It's based on [charms](https://juju.ubuntu.com/docs/charm-store.html) and very
straightforward to use. Here is a very basic set of commands with which you can
deploy a Wordpress related to a MySQL service:

```
% juju bootstrap
% juju deploy mysql
% juju deploy wordpress
% juju add-relation wordpress mysql
% juju expose wordpress
```

These commands will boostrap the environment, setting up a bootstrap machine
which will manage your services; deploy mysql and wordpress instances; add a
relation between them; and expose the wordpress port. The voilà, we have a
wordpress deployed, and ready to serve our posts. Amazing, huh?

But there is an issue: although you can install the `juju` command line tool in
almost any OS (including Mac OS), right now you are able do deploy only
Ubuntu-based services (you must use an Ubuntu instance or container).

To change this behavior, and enable Juju to spawn CentOS instances (and
containers, if you have a CentOS lxc template), we need to develop and apply
some changes to Juju and
[cloud-init](https://help.ubuntu.com/community/CloudInit). Juju uses cloud-init
to spawn machines with proper dependencies set up, and it's based on modules.
All we need to do, is add a module able to install rpm packages using `yum`.

`cloud-init` modules are Python modules that starts with `cc_` and implement a
`handle` function (for example, a module called "yum_packages" would be written
to a file called `cc_yum_packages.py`). So, here is the code for the module
`yum_packages`:

```python
import subprocess
import traceback

from cloudinit import CloudConfig, util

frequency = CloudConfig.per_instance

def yum_install(packages):
    cmd = ["yum", "--quiet", "--assumeyes", "install"]
    cmd.extend(packages)
    subprocess.check_call(cmd)

def handle(_name, cfg, _cloud, log, args):
    pkglist = util.get_cfg_option_list_or_str(cfg, "packages", [])

    if pkglist:
        try:
            yum_install(pkglist)
        except subprocess.CalledProcessError:
            log.warn("Failed to install yum packages: %s" % pkglist)
            log.debug(traceback.format_exc())
            raise

    return True
```

The module installs all packages listed in cloud-init yaml file. If we want to install `emacs-nox` package, we would write this yaml file and use it as user data in the instance:

```
#cloud-config
modules:
 - yum_packages
packages: [emacs-nox]
```

`cloud-init` already works on Fedora, with Python 2.7, but to work on CentOS 6, with Python 2.6, it needs a patch:

```
--- cloudinit/util.py 2012-05-22 12:18:21.000000000 -0300
+++ cloudinit/util.py 2012-05-31 12:44:24.000000000 -0300
@@ -227,7 +227,7 @@
         stderr=subprocess.PIPE, stdin=subprocess.PIPE)
     out, err = sp.communicate(input_)
     if sp.returncode is not 0:
-        raise subprocess.CalledProcessError(sp.returncode, args, (out, err))
+        raise subprocess.CalledProcessError(sp.returncode, args)
     return(out, err)
```

I've packet up this module and this patch in a [RPM
package](https://github.com/globocom/cloudinit-centos-6) that must be
pre-installed in the lxc template and AMI images. Now, we need to change Juju
in order to make it use the `yum_packages` module, and include all RPM packages
that we need to install when the machine borns.

Is Juju, there is a class that is responsible for building and rendering the
YAML file used by cloud-init. We can extend it and change only two methods:
`_collect_packages`, that returns the list of packages that will be installed
in the machine after it is spawned; and `render` that returns the file itself.
Here is our `CentOSCloudInit` class (within the patch):

```
diff -u juju-0.5-bzr531.orig/juju/providers/common/cloudinit.py juju-0.5-bzr531/juju/providers/common/cloudinit.py
--- juju-0.5-bzr531.orig/juju/providers/common/cloudinit.py 2012-05-31 15:42:17.480769486 -0300
+++ juju-0.5-bzr531/juju/providers/common/cloudinit.py 2012-05-31 15:55:13.342884919 -0300
@@ -324,3 +324,32 @@
             "machine-id": self._machine_id,
             "juju-provider-type": self._provider_type,
             "juju-zookeeper-hosts": self._join_zookeeper_hosts()}
+
+
+class CentOSCloudInit(CloudInit):
+
+    def _collect_packages(self):
+        packages = [
+            "bzr", "byobu", "tmux", "python-setuptools", "python-twisted",
+            "python-txaws", "python-zookeeper", "python-devel", "juju"]
+        if self._zookeeper:
+            packages.extend([
+                "zookeeper", "libzookeeper", "libzookeeper-devel"])
+        return packages
+
+    def render(self):
+        """Get content for a cloud-init file with appropriate specifications.
+
+        :rtype: str
+
+        :raises: :exc:`juju.errors.CloudInitError` if there isn't enough
+            information to create a useful cloud-init.
+        """
+        self._validate()
+        return format_cloud_init(
+            self._ssh_keys,
+            packages=self._collect_packages(),
+            repositories=self._collect_repositories(),
+            scripts=self._collect_scripts(),
+            data=self._collect_machine_data(),
+            modules=["ssh", "yum_packages", "runcmd"])
```

The other change we need is in the `format_cloud_init` function, in order to
make it recognize the `modules` parameter that we used above, and tell
cloud-init to not run `apt-get` (update nor upgrade). Here is the patch:

```
diff -ur juju-0.5-bzr531.orig/juju/providers/common/utils.py juju-0.5-bzr531/juju/providers/common/utils.py
--- juju-0.5-bzr531.orig/juju/providers/common/utils.py 2012-05-31 15:42:17.480769486 -0300
+++ juju-0.5-bzr531/juju/providers/common/utils.py 2012-05-31 15:44:06.605014021 -0300
@@ -85,7 +85,7 @@

 def format_cloud_init(
-    authorized_keys, packages=(), repositories=None, scripts=None, data=None):
+    authorized_keys, packages=(), repositories=None, scripts=None, data=None, modules=None):
     """Format a user-data cloud-init file.

     This will enable package installation, and ssh access, and script
@@ -117,8 +117,8 @@
         structure.
     """
     cloud_config = {
-        "apt-update": True,
-        "apt-upgrade": True,
+        "apt-update": False,
+        "apt-upgrade": False,
         "ssh_authorized_keys": authorized_keys,
         "packages": [],
         "output": {"all": "| tee -a /var/log/cloud-init-output.log"}}
@@ -136,6 +136,11 @@
     if scripts:
         cloud_config["runcmd"] = scripts

+    if modules:
+        cloud_config["modules"] = modules
+
     output = safe_dump(cloud_config)
     output = "#cloud-config\n%s" % (output)
     return output
```

This patch is also packed up within
[juju-centos-6](https://github.com/globocom/juju-centos-6) repository, which
provides sources for building RPM packages for juju, and also some pre-built
RPM packages.

Now just build an AMI image with `cloudinit` pre-installed, configure your juju
`environments.yaml` file to use this image in the environment and you are ready
to deploy cloud services on CentOS machines using Juju!

Some caveats:

- Juju needs a user called `ubuntu` to interact with its machines, so you will
  need to create this user in your CentOS AMI/template.
- You need to host all RPM packages for
  [juju](https://github.com/globocom/juju-centos-6),
  [cloud-init](https://github.com/globocom/cloudinit-centos-6) and following
  dependencies in some `yum` repository (I haven't submitted them to any public
  repository):
    - [python-txaws](https://github.com/globocom/python-txaws-centos-6)
    - [python-txzookeeper](https://github.com/globocom/python-txzookeeper-centos-6)
    - [zookeeper](https://github.com/globocom/zookeeper-centos-6)
- With this patched Juju, you will have a pure-centos cloud. It does not enable
  you to have multiple OSes in the same environment.

It's important to notice that we are going to put some effort to make the Go
version of juju born supporting multiple OSes, ideally through an interface
that makes it extensible to any other OS, not Ubuntu and CentOS only.
