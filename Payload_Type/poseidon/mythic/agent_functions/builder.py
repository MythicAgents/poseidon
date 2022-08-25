from mythic_payloadtype_container.PayloadBuilder import *
from mythic_payloadtype_container.MythicCommandBase import *
import asyncio
import os
import shutil
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
            name="architecture",
            parameter_type=BuildParameterType.ChooseOne,
            description="Choose the agent's architecture ",
            choices=["AMD_x64", "ARM_x64"],
            default_value="AMD_x64",
        ),
        BuildParameter(
            name="proxy_bypass",
            parameter_type=BuildParameterType.Boolean,
            default_value=False,
            description="Ignore HTTP proxy environment settings configured on the target host?",
        ),
        BuildParameter(
            name="garble",
            description="Use Garble to obfuscate the output Go executable. "
            "\nWARNING - This significantly slows the agent build time.",
            parameter_type=BuildParameterType.Boolean,
            default_value=True,
            required=False,
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
        elif self.selected_os == "Windows":
            target_os = "windows"
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

            ldflags += " -X '{}.proxy_bypass={}'".format(
                poseidon_repo_profile, self.get_parameter("proxy_bypass")
            )
            # Set the Go -buildid argument to an empty string to remove the indicator
            ldflags += " -buildid="
            goarch = "amd64"
            if self.get_parameter("architecture") == "ARM_x64":
                goarch = "arm64"
            command = f"rm -rf /build; rm -rf /deps; CGO_ENABLED=1 GOOS={target_os} GOARCH={goarch} "

            go_cmd = f'-tags {profile} -buildmode {self.get_parameter("mode")} -ldflags "{ldflags}"'
            if target_os == "darwin":
                command += "CC=o64-clang CXX=o64-clang++ "
            elif target_os == "windows":
                command += "CC=x86_64-w64-mingw32-gcc "
            # command += "export GOGARBLE=golang.org,github.com,howett.net;"
            # command += "export GOGARBLE=$GOGARBLE,vendor,net,internal,reflect,crypto,strings,math,compress,compress,syscall,os,unicode,context,regexp,sync,strconv,sort,fmt,bytes,path,bufio,log,mime,hash,container;"
            command += "GOGARBLE=* "
            if self.get_parameter("garble"):
                command += (
                    "/go/src/bin/garble -tiny -literals -debug -seed random build "
                )
            else:
                command += "go build "
                # This shouldn't be necessary
                # Don't include encoding
            command += f"{go_cmd} -o /build/poseidon-{target_os}"
            if target_os == "darwin":
                command += f"-{macOSVersion}"
            command += f"-{goarch}"
            if self.get_parameter("mode") == "c-shared":
                if target_os == "windows":
                    command += ".dll"
                elif target_os == "darwin":
                    command += ".dylib"
                else:
                    command += ".so"
            elif self.get_parameter("mode") == "c-archive":
                command += ".a"

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
                    resp.build_message += f"\n[BUILD]{command}\n"
            if stderr:
                resp.build_stderr += f"\n[STDERR]\n{stderr.decode()}"
                if debug:
                    resp.build_stderr += f"\n[BUILD]{command}\n"

            # default build mode
            if self.get_parameter("mode") == "default":
                # Linux
                if target_os == "linux" or target_os == "windows":
                    if os.path.exists(f"/build/poseidon-{target_os}-{goarch}"):
                        resp.payload = open(
                            f"/build/poseidon-{target_os}-{goarch}", "rb"
                        ).read()
                    else:
                        resp.build_stderr += (
                            f"/build/poseidon-{target_os}-{goarch} does not exist"
                        )
                        resp.status = BuildStatus.Error
                        return resp
                # Darwin (macOS)
                elif target_os == "darwin":
                    if os.path.exists(
                        f"/build/poseidon-{target_os}-{macOSVersion}-{goarch}"
                    ):
                        resp.payload = open(
                            f"/build/poseidon-{target_os}-{macOSVersion}-{goarch}", "rb"
                        ).read()
                    else:
                        resp.build_stderr += f"/build/poseidon-{target_os}-{macOSVersion}-{goarch} does not exist"
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
                    if os.path.exists(f"/build/poseidon-{target_os}-{goarch}.so"):
                        resp.payload = open(
                            f"/build/poseidon-{target_os}-{goarch}.so", "rb"
                        ).read()
                    else:
                        resp.build_stderr += (
                            f"/build/poseidon-{target_os}-{goarch}.so does not exist"
                        )
                        resp.status = BuildStatus.Error
                        return resp
                # Darwin (macOS)
                elif target_os == "darwin":
                    if os.path.exists(
                        f"/build/poseidon-{target_os}-{macOSVersion}-{goarch}.dylib"
                    ):
                        resp.payload = open(
                            f"/build/poseidon-{target_os}-{macOSVersion}-{goarch}.dylib",
                            "rb",
                        ).read()
                    else:
                        resp.build_stderr += f"/build/poseidon-{target_os}-{macOSVersion}-{goarch}.dylib does not exist"
                        resp.status = BuildStatus.Error
                        return resp
                elif target_os == "windows":
                    if os.path.exists(f"/build/poseidon-{target_os}-{goarch}.dll"):
                        resp.payload = open(
                            f"/build/poseidon-{target_os}-{goarch}.dll", "rb"
                        ).read()
                    else:
                        resp.build_stderr += (
                            f"/build/poseidon-{target_os}-{goarch}.dll does not exist"
                        )
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
                    if os.path.exists(f"/build/poseidon-{target_os}-{goarch}.a"):
                        shutil.make_archive(
                            f"{agent_build_path}/poseidon", "zip", "/build"
                        )
                        resp.payload = open(
                            f"{agent_build_path}/poseidon" + ".zip", "rb"
                        ).read()
                    else:
                        resp.build_stderr += (
                            f"/build/poseidon-{target_os}-{goarch}.a does not exist"
                        )
                        resp.status = BuildStatus.Error
                        return resp
                # Darwin (macOS)
                elif target_os == "darwin":
                    if os.path.exists(
                        f"/build/poseidon-{target_os}-{macOSVersion}-{goarch}.a"
                    ):
                        shutil.make_archive(
                            f"{agent_build_path}/poseidon", "zip", "/build"
                        )
                        resp.payload = open(
                            f"{agent_build_path}/poseidon" + ".zip", "rb"
                        ).read()
                    else:
                        resp.build_stderr += f"/build/poseidon-{target_os}-{macOSVersion}-{goarch}.a does not exist"
                        resp.status = BuildStatus.Error
                        return resp
                else:
                    resp.build_stderr += (
                        f"Unhandled operating system: {target_os} for "
                        f"{self.get_parameter('mode')} build mode"
                    )
                    resp.status = BuildStatus.Error
                    return resp
            # Unhandled
            else:
                resp.build_stderr += (
                    f"Unhandled build mode {self.get_parameter('mode')}"
                )
                resp.status = BuildStatus.Error
                return resp

            # Successfully created the payload without error
            resp.build_message += (
                f"\nCreated Poseidon payload!\n"
                f"OS: {target_os}, "
                f'Build Mode: {self.get_parameter("mode")}, '
                f"C2 Profile: {profile}\n[BUILD]{command}\n"
            )
            resp.status = BuildStatus.Success
            return resp
        except Exception as e:
            resp.build_stderr += "\n" + str(e)
        return resp
