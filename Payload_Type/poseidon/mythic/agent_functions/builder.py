from mythic_payloadtype_container.PayloadBuilder import *
from mythic_payloadtype_container.MythicCommandBase import *
import asyncio
import os
import shutil
import json

# Enable additional message details to the Mythic UI
debug = True


class Poseidon(PayloadType):
    name = "poseidon"
    file_extension = "bin"
    author = "@xorrior, @djhohnstein, @Ne0nd0g"
    supported_os = [SupportedOS.Linux, SupportedOS.MacOS]
    wrapper = False
    wrapped_payloads = []
    note = "A fully featured macOS and Linux Golang agent"
    supports_dynamic_loading = False
    mythic_encrypts = True
    build_parameters = [
        BuildParameter(
            name="mode",
            parameter_type=BuildParameterType.ChooseOne,
            description="Choose the build mode option. Select default for executables, "
                        "c-shared for a .dylib or .so file, "
                        "or c-archive for a .Zip containing C source code with an archive and header file",
            choices=["default", "c-archive", "c-shared"],
            default_value="default",
        ),
        BuildParameter(
            name="proxy_bypass",
            parameter_type=BuildParameterType.Boolean,
            default_value=False,
            description="Ignore HTTP proxy environment settings configured on the target host?",
        ),
    ]
    c2_profiles = ["websocket", "http", "poseidon_tcp"]

    async def build(self) -> BuildResponse:
        macOSVersion = "10.12"
        # this function gets called to create an instance of your payload
        resp = BuildResponse(status=BuildStatus.Error)
        target_os = "linux"
        if self.selected_os == "macOS":
            target_os = "darwin"
        if len(self.c2info) != 1:
            resp.build_stderr = "Poseidon only accepts one c2 profile at a time"
            return resp
        try:
            agent_build_path = "/Mythic/agent_code"

            # Get the selected C2 profile information (e.g., http or websocket)
            c2 = self.c2info[0]
            profile = c2.get_c2profile()["name"]
            if profile not in self.c2_profiles:
                resp.build_message = "Invalid c2 profile name specified"
                return resp

            # This package path is used with Go's "-X" link flag to set the value string variables in code at compile
            # time. This is how each profile's configurable options are passed in.
            poseidon_repo_profile = f"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"

            # Build Go link flags that are passed in at compile time through the "-ldflags=" argument
            # https://golang.org/cmd/link/
            ldflags = f"-s -w -X '{poseidon_repo_profile}.UUID={self.uuid}'"

            # Iterate over the C2 profile parameters and associated variable through Go's "-X" link flag
            for key, val in c2.get_parameters_dict().items():
                # dictionary instances will be crypto components
                if isinstance(val, dict):
                    ldflags += f" -X '{poseidon_repo_profile}.{key}={val['enc_key']}'"
                elif key == "headers":
                    v = json.dumps(val).replace('"', '\\"')
                    ldflags += f" -X '{poseidon_repo_profile}.{key}={v}'"
                else:
                    if val:
                        ldflags += f" -X '{poseidon_repo_profile}.{key}={val}'"

            ldflags += " -X '{}.proxy_bypass={}'".format(poseidon_repo_profile, self.get_parameter("proxy_bypass"))
            # Set the Go -buildid argument to an empty string to remove the indicator
            ldflags += " -buildid="
            command = "rm -rf /build; rm -rf /deps;"
            command += (
                "xgo -tags={} --targets={}/{} -buildmode={} -ldflags=\"{}\" -out poseidon .".format(
                    profile,
                    target_os,
                    "amd64",
                    self.get_parameter("mode"),
                    ldflags,
                )
            )

            # Execute the constructed xgo command to build Poseidon
            proc = await asyncio.create_subprocess_shell(
                command,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                cwd=agent_build_path,
            )

            # Collect and data written Standard Output and Standard Error
            stdout, stderr = await proc.communicate()
            if stdout:
                resp.build_stdout += f"\n[STDOUT]\n{stdout.decode()}"
                if debug:
                    resp.build_message += f'\n[BUILD]{command}\n'
            if stderr:
                resp.build_stderr += f"\n[STDERR]\n{stderr.decode()}"
                if debug:
                    resp.build_stderr += f'\n[BUILD]{command}\n'

            # default build mode
            if self.get_parameter("mode") == "default":
                # Linux
                if target_os == "linux":
                    if os.path.exists(f"/build/poseidon-{target_os}-amd64"):
                        resp.payload = open(f"/build/poseidon-{target_os}-amd64", "rb").read()
                    else:
                        resp.build_stderr += f"/build/poseidon-{target_os}-amd64 does not exist"
                        resp.status = BuildStatus.Error
                        return resp
                # Darwin (macOS)
                elif target_os == "darwin":
                    if os.path.exists(f"/build/poseidon-{target_os}-{macOSVersion}-amd64"):
                        resp.payload = open(f"/build/poseidon-{target_os}-{macOSVersion}-amd64", "rb").read()
                    else:
                        resp.build_stderr += f"/build/poseidon-{target_os}-{macOSVersion}-amd64 does not exist"
                        resp.status = BuildStatus.Error
                        return resp
                else:
                    resp.build_stderr += f"Unhandled operating system: {target_os} for {self.get_parameter('mode')} build mode"
                    resp.status = BuildStatus.Error
                    return resp
            # C-shared (e.g., Dylib or SO)
            elif self.get_parameter("mode") == "c-shared":
                # Linux
                if target_os == "linux":
                    if os.path.exists(f"/build/poseidon-{target_os}-amd64.so"):
                        resp.payload = open(f"/build/poseidon-{target_os}-amd64.so", "rb").read()
                    else:
                        resp.build_stderr += f"/build/poseidon-{target_os}-amd64.so does not exist"
                        resp.status = BuildStatus.Error
                        return resp
                # Darwin (macOS)
                elif target_os == "darwin":
                    if os.path.exists(f"/build/poseidon-{target_os}-{macOSVersion}-amd64.dylib"):
                        resp.payload = open(f"/build/poseidon-{target_os}-{macOSVersion}-amd64.dylib",
                                            "rb").read()
                    else:
                        resp.build_stderr += f"/build/poseidon-{target_os}-{macOSVersion}-amd64.dylib does not exist"
                        resp.status = BuildStatus.Error
                        return resp
                else:
                    resp.build_stderr += f"Unhandled operating system: {target_os} for {self.get_parameter('mode')} build mode"
                    resp.status = BuildStatus.Error
                    return resp
            # C-shared (e.g., Dylib or SO)
            elif self.get_parameter("mode") == "c-archive":
                # Copy the C file into the build directory
                file1 = open(
                    f"/Mythic/agent_code/sharedlib/sharedlib-darwin-linux.c", "r"
                ).read()
                with open("/build/sharedlib-darwin-linux.c", "w") as f:
                    f.write(file1)
                # Linux
                if target_os == "linux":
                    if os.path.exists(f"/build/poseidon-{target_os}-amd64.a"):
                        shutil.make_archive(f"{agent_build_path}/poseidon", "zip", "/build")
                        resp.payload = open(f"{agent_build_path}/poseidon" + ".zip", "rb").read()
                    else:
                        resp.build_stderr += f"/build/poseidon-{target_os}-amd64.a does not exist"
                        resp.status = BuildStatus.Error
                        return resp
                # Darwin (macOS)
                elif target_os == "darwin":
                    if os.path.exists(f"/build/poseidon-{target_os}-{macOSVersion}-amd64.a"):
                        shutil.make_archive(f"{agent_build_path}/poseidon", "zip", "/build")
                        resp.payload = open(f"{agent_build_path}/poseidon" + ".zip", "rb").read()
                    else:
                        resp.build_stderr += f"/build/poseidon-{target_os}-10.06-amd64.a does not exist"
                        resp.status = BuildStatus.Error
                        return resp
                else:
                    resp.build_stderr += f"Unhandled operating system: {target_os} for " \
                                         f"{self.get_parameter('mode')} build mode"
                    resp.status = BuildStatus.Error
                    return resp
            # Unhandled
            else:
                resp.build_stderr += f"Unhandled build mode {self.get_parameter('mode')}"
                resp.status = BuildStatus.Error
                return resp

            # Successfully created the payload without error
            resp.build_message += f'\nCreated Poseidon payload!\n' \
                                  f'OS: {target_os}, ' \
                                  f'Build Mode: {self.get_parameter("mode")}, ' \
                                  f'C2 Profile: {profile}\n'
            resp.status = BuildStatus.Success
            return resp
        except Exception as e:
            resp.build_stderr += "\n" + str(e)
        return resp
