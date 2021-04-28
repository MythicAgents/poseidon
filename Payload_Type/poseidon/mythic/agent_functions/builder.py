from mythic_payloadtype_container.PayloadBuilder import *
from mythic_payloadtype_container.MythicCommandBase import *
import asyncio
import os
import shutil
import uuid
import json

# Enable additional message details to the Mythic UI
debug = False


class Poseidon(PayloadType):
    name = "poseidon"
    file_extension = "bin"
    author = "@xorrior, @djhohnstein, @Ne0nd0g"
    supported_os = [SupportedOS.Linux, SupportedOS.MacOS]
    wrapper = False
    wrapped_payloads = []
    note = "A fully featured macOS and Linux Golang agent"
    supports_dynamic_loading = False
    build_parameters = {
        "os": BuildParameter(
            name="os",
            parameter_type=BuildParameterType.ChooseOne,
            description="Choose the target OS",
            choices=["darwin", "linux"],
            default_value="linux",
        ),
        "mode": BuildParameter(
            name="mode",
            parameter_type=BuildParameterType.ChooseOne,
            description="Choose the buildmode option. Default for executables, c-archive/archive/c-shared/shared are for libraries",
            choices=["default", "c-archive"],
            default_value="default",
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
            resp.message = "Poseidon only accepts one c2 profile at a time"
            return resp
        try:
            c2 = self.c2info[0]
            profile = c2.get_c2profile()["name"]
            if profile not in self.c2_profiles:
                resp.message = "Invalid c2 profile name specified"
                return resp

            poseidon_repo_profile = f"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/profiles/{profile}"
            ldflags = f"-s -w -X 'main.c2Profile={profile}' -X '{poseidon_repo_profile}.UUID={self.uuid}'"

            for key, val in c2.get_parameters_dict().items():
                if isinstance(val, dict):
                    ldflags += f" -X '{poseidon_repo_profile}.{key}={val['enc_key']}'"
                elif key == "headers":
                    v = json.dumps(val).replace('"', '\\"')
                    ldflags += f" -X '{poseidon_repo_profile}.{key}={v}'"
                else:
                    if val:
                        ldflags += f" -X '{poseidon_repo_profile}.{key}={val}'"

            ldflags += " -buildid="
            command = "rm -rf /build; rm -rf /deps;"
            command += (
                "xgo -tags={} --targets={}/{} -buildmode={} -ldflags=\"{}\" -out poseidon .".format(
                    profile,
                    "darwin" if self.get_parameter("os") == "darwin" else "linux",
                    "amd64",
                    "default" if self.get_parameter("mode") == "default" else "c-archive",
                    ldflags,
                )
            )
            proc = await asyncio.create_subprocess_shell(
                command,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                cwd="/Mythic/agent_code",
            )
            stdout, stderr = await proc.communicate()
            if stdout:
                resp.message = f"[stdout]\n{stdout.decode()}"
            if stderr:
                resp.build_error += f"[stderr]\n{stderr.decode()}"
                if debug:
                    resp.build_error += f'\n[BUILD]{command}\n'
            if os.path.exists("/build"):
                files = os.listdir("/build")
                if len(files) == 1:
                    if self.get_parameter('os') == "darwin":
                        resp.payload = open(f"/build/poseidon-{self.get_parameter('os')}-10.06-amd64", "rb").read()
                    else:
                        resp.payload = open(f"/build/poseidon-{self.get_parameter('os')}-amd64", "rb").read()
                    resp.message += "\nCreated payload!\n"
                else:
                    temp_uuid = str(uuid.uuid4())
                    file1 = open(
                        f"/Mythic/agent_code/sharedlib/sharedlib-darwin-linux.c", "r"
                    ).read()
                    with open("/build/sharedlib-darwin-linux.c", "w") as f:
                        f.write(file1)
                    shutil.make_archive(f"/Mythic/agent_code/{temp_uuid}", "zip", "/build")
                    resp.payload = open(f"/Mythic/agent_code/{temp_uuid}" + ".zip", "rb").read()
                    resp.message = "Created a zip archive of files!\n"
                resp.status = BuildStatus.Success
            if debug:
                resp.message += f'\n[BUILD]{command}\n'
            else:
                # something went wrong, return our errors
                resp.build_error += "\nNo files created"
        except Exception as e:
            resp.build_error += "\n" + str(e)
        return resp
