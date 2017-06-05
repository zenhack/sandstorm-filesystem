import base64
import os
import subprocess
import sys

sandstorm_repo = sys.argv[1]
schema_file = sys.argv[2]

with open(schema_file) as f:
    proc = subprocess.Popen(['capnp', 'encode', '-p',
                             os.path.join(sandstorm_repo,
                                          'src/sandstorm/powerbox.capnp'),
                             'PowerboxDescriptor',
                             ], stdin=f, stdout=subprocess.PIPE)
    out, _ = proc.communicate()
    sys.stdout.write(base64.urlsafe_b64encode(out).strip('='))
