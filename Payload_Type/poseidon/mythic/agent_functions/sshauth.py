from mythic_payloadtype_container.MythicCommandBase import *
import json


class SSHAuthArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="username",
                type=ParameterType.String,
                description="Authenticate to the designated hosts using this username.",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=1,
                        group_name="scp-private-key"
                    ),
                    ParameterGroupInfo(
                        group_name="scp-plaintext-password"
                    ),
                    ParameterGroupInfo(
                        group_name="run-command-plaintext-password"
                    ),
                    ParameterGroupInfo(
                        group_name="run-command-private-key"
                    )
                ]
            ),
            CommandParameter(
                name="source",
                type=ParameterType.String,
                description="If doing SCP, this is the source file",
                default_value="",
                parameter_group_info=[
                    ParameterGroupInfo(
                        group_name="scp-private-key"
                    ),
                    ParameterGroupInfo(
                        group_name="scp-plaintext-password"
                    )
                ]
            ),
            CommandParameter(
                name="destination",
                type=ParameterType.String,
                description="If doing SCP, this is the destination file",
                default_value="",
                parameter_group_info=[
                    ParameterGroupInfo(
                        group_name="scp-private-key"
                    ),
                    ParameterGroupInfo(
                        group_name="scp-plaintext-password"
                    )
                ]
            ),
            CommandParameter(
                name="private_key",
                type=ParameterType.String,
                description="Authenticate to the designated hosts using this private key",
                parameter_group_info=[
                    ParameterGroupInfo(
                        group_name="scp-private-key"
                    ),
                    ParameterGroupInfo(
                        group_name="command-private-key"
                    )
                ]
            ),
            CommandParameter(
                name="port",
                type=ParameterType.Number,
                description="SSH Port if different than 22",
                default_value="22",
                parameter_group_info=[
                    ParameterGroupInfo(
                        required=False,
                        group_name="scp-private-key"
                    ),
                    ParameterGroupInfo(
                        required=False,
                        group_name="scp-plaintext-password"
                    ),
                    ParameterGroupInfo(
                        group_name="run-command-plaintext-password"
                    ),
                    ParameterGroupInfo(
                        group_name="run-command-private-key"
                    )
                ]
            ),
            CommandParameter(
                name="password",
                type=ParameterType.String,
                description="Authenticate to the designated hosts using this password",
                default_value="",
                parameter_group_info=[
                    ParameterGroupInfo(
                        group_name="scp-plaintext-password"
                    ),
                    ParameterGroupInfo(
                        group_name="command-plaintext-password"
                    )
                ]
            ),
            CommandParameter(
                name="hosts",
                type=ParameterType.Array,
                description="Hosts that you will auth to",
                parameter_group_info=[
                    ParameterGroupInfo(
                        group_name="scp-plaintext-password"
                    ),
                    ParameterGroupInfo(
                        group_name="scp-private-key"
                    ),
                    ParameterGroupInfo(
                        group_name="run-command-plaintext-password"
                    ),
                    ParameterGroupInfo(
                        group_name="run-command-private-key"
                    )
                ]
            ),
            CommandParameter(
                name="command",
                type=ParameterType.String,
                description="Command to execute on remote systems if not doing SCP",
                default_value="",
                parameter_group_info=[
                    ParameterGroupInfo(
                        group_name="run-command-plaintext-password"
                    ),
                    ParameterGroupInfo(
                        group_name="run-command-private-key"
                    )
                ]
            ),
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


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
        if "private-key" in task.args.get_parameter_group_name():
            task.display_params += " with a private key "
        else:
            task.display_params += " with password, " + task.args.get_arg("password") + ", "
        if "command" in task.args.get_parameter_group_name():
            task.display_params += "to run command, " + task.args.get_arg("command")
        elif "scp" in task.args.get_parameter_group_name():
            task.display_params += "to copy local file, " + task.args.get_arg("source") + ", to " + task.args.get_arg("destination")
        return task

    async def process_response(self, response: AgentResponse):
        pass
