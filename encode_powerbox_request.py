import base64
import os
import subprocess
import sys

sandstorm_repo = sys.argv[1]

with open('ro-dir-powerbox-request.capnp') as f:
    proc = subprocess.Popen(['capnp', 'encode', '-p',
                             os.path.join(sandstorm_repo,
                                          'src/sandstorm/powerbox.capnp'),
                             'PowerboxDescriptor',
                             ], stdin=f, stdout=subprocess.PIPE)
    out, _ = proc.communicate()
    sys.stdout.write(base64.urlsafe_b64encode(out).strip('='))
