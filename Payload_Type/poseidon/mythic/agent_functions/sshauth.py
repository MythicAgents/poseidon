from mythic_payloadtype_container.MythicCommandBase import *
import json


class SSHAuthArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {
            "username": CommandParameter(
                name="username",
                type=ParameterType.String,
                description="Authenticate to the designated hosts using this username.",
                ui_position=1
            ),
            "source": CommandParameter(
                name="source",
                type=ParameterType.String,
                description="If doing SCP, this is the source file",
                required=False,
                default_value="",
                ui_position=7
            ),
            "destination": CommandParameter(
                name="destination",
                type=ParameterType.String,
                description="If doing SCP, this is the destination file",
                required=False,
                default_value="",
                ui_position=8
            ),
            "private_key": CommandParameter(
                name="private_key",
                type=ParameterType.String,
                description="Authenticate to the designated hosts using this private key",
                required=False,
                ui_position=3
            ),
            "port": CommandParameter(
                name="port",
                type=ParameterType.Number,
                description="SSH Port if different than 22",
                default_value="22",
                ui_position=5
            ),
            "password": CommandParameter(
                name="password",
                type=ParameterType.String,
                description="Authenticate to the designated hosts using this password",
                required=False,
                default_value="",
                ui_position=2
            ),
            "hosts": CommandParameter(
                name="hosts",
                type=ParameterType.Array,
                description="Hosts that you will auth to",
                ui_position=4
            ),
            "command": CommandParameter(
                name="command",
                type=ParameterType.String,
                description="Command to execute on remote systems if not doing SCP",
                required=False,
                default_value="",
                ui_position=6
            ),
        }

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)


class SSHAuthCommand(CommandBase):
    cmd = "sshauth"
    needs_admin = False
    help_cmd = "sshauth"
    description = """SSH to specified host(s) using the designated credentials. 
You can also use this to execute a specific command on the remote hosts via SSH or use it to SCP files.
"""
    version = 1
    author = "@xorrior"
    argument_class = SSHAuthArguments
    attackmapping = ["T1110"]
    browser_script = BrowserScript(script_name="sshauth", author="@djhohnstein")

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        task.display_params = "Authenticate as " + task.args.get_arg("username")
        if task.args.get_arg("private_key") != "":
            task.display_params += " with a private key "
        else:
            task.display_params += " with password, " + task.args.get_arg("password") + ", "
        if task.args.get_arg("command") != "":
            task.display_params += "to run command, " + task.args.get_arg("command")
        elif task.args.get_arg("source") != "":
            task.display_params += "to copy local file, " + task.args.get_arg("source") + ", to " + task.args.get_arg("destination")
        return task

    async def process_response(self, response: AgentResponse):
        pass
