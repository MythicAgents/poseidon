from mythic_payloadtype_container.PayloadBuilder import *
from mythic_payloadtype_container.MythicCommandBase import *
import asyncio
import os
import tempfile
from distutils.dir_util import copy_tree
import shutil
import uuid
import json
import sys


class Poseidon(PayloadType):

    name = "poseidon"
    file_extension = "bin"
    author = "@xorrior, @djhohnstein"
    supported_os = [SupportedOS.Linux, SupportedOS.MacOS]
    wrapper = False
    wrapped_payloads = []
    note = "A fully featured macOS and Linux Golang agent"
    supports_dynamic_loading = False
    mythic_encrypts = True
    build_parameters = {
        "mode": BuildParameter(
            name="mode",
            parameter_type=BuildParameterType.ChooseOne,
            description="Choose the buildmode option. Default for executables, c-archive/archive/c-shared/shared are for libraries",
            choices=["default", "c-archive"],
        ),
    }
    c2_profiles = ["websocket", "http"]
    support_browser_scripts = [
        BrowserScript(script_name="create_table", author="@djhohnstein"),
        BrowserScript(script_name="collapsable", author="@djhohnstein"),
        BrowserScript(
            script_name="file_size_to_human_readable_string", author="@djhohnstein"
        ),
        BrowserScript(
            script_name="create_process_additional_info_modal", author="@djhohnstein"
        ),
        BrowserScript(
            script_name="show_process_additional_info_modal", author="@djhohnstein"
        ),
        BrowserScript(
            script_name="copy_additional_info_to_clipboard", author="@djhohnstein"
        ),
    ]

    async def build(self) -> BuildResponse:
        # this function gets called to create an instance of your payload
        resp = BuildResponse(status=BuildStatus.Error)
        if len(self.c2info) != 1:
            resp.build_stderr = "Poseidon only accepts one c2 profile at a time"
            return resp
        try:
            agent_build_path = tempfile.TemporaryDirectory(suffix=self.uuid).name
            # shutil to copy payload files over
            copy_tree(self.agent_code_path, agent_build_path)
            file1 = open(
                "{}/pkg/profiles/profile.go".format(agent_build_path), "r"
            ).read()
            file1 = file1.replace("UUID_HERE", self.uuid)
            with open("{}/pkg/profiles/profile.go".format(agent_build_path), "w") as f:
                f.write(file1)
            c2 = self.c2info[0]
            profile = c2.get_c2profile()["name"]
            if profile not in self.c2_profiles:
                resp.build_message = "Invalid c2 profile name specified"
                return resp
            file1 = open(
                "{}/c2_profiles/{}.go".format(agent_build_path, profile), "r"
            ).read()
            for key, val in c2.get_parameters_dict().items():
                # dictionary instances will be crypto components
                if isinstance(val, dict):
                    file1 = file1.replace(key, val["enc_key"] if val["enc_key"] is not None else "")
                # headers is the name of the c2 profile parameter with an array of header values to set
                elif key == "headers":
                    file1 = file1.replace(key, json.dumps(json.dumps(val)))
                else:
                    file1 = file1.replace(key, val)
            with open(
                "{}/pkg/profiles/{}.go".format(agent_build_path, profile), "w"
            ) as f:
                f.write(file1)
            command = "rm -rf /build; rm -rf /deps; rm -rf /go/src/poseidon;"
            command += "mkdir -p /go/src/poseidon/src; mv * /go/src/poseidon/src; mv /go/src/poseidon/src/poseidon.go /go/src/poseidon/;"
            command += "cd /go/src/poseidon;"
            command += (
                "xgo -tags={} --targets={}/{} -buildmode={} -out poseidon .".format(
                    profile,
                    "darwin" if self.selected_os == "macOS" else "linux",
                    "amd64",
                    "default" if self.get_parameter("mode") == "default" else "c-archive",
                )
            )
            proc = await asyncio.create_subprocess_shell(
                command,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                cwd=agent_build_path,
            )
            stdout, stderr = await proc.communicate()
            if stdout:
                resp.build_message = f"[stdout]\n{stdout.decode()}"
            if stderr:
                resp.build_stderr += f"[stderr]\n{stderr.decode()}"
            if os.path.exists("/build"):
                files = os.listdir("/build")
                if len(files) == 1:
                    resp.payload = open("/build/" + files[0], "rb").read()
                    resp.build_message = "\nCreated payload!\n" + resp.build_message
                else:
                    temp_uuid = str(uuid.uuid4())
                    file1 = open(
                        "/go/src/poseidon/src/sharedlib-darwin-linux.c", "r"
                    ).read()
                    with open("/build/sharedlib-darwin-linux.c", "w") as f:
                        f.write(file1)
                    shutil.make_archive(f"{agent_build_path}/{temp_uuid}", "zip", "/build")
                    resp.payload = open(f"{agent_build_path}/{temp_uuid}" + ".zip", "rb").read()
                    resp.build_message = "Created a zip archive of files!\n" + resp.build_message
                resp.status = BuildStatus.Success
            else:
                # something went wrong, return our errors
                resp.build_stderr += "\nNo files created"
        except Exception as e:
            resp.build_stderr += "\n" + str(e)
        return resp
