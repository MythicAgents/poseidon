from mythic_payloadtype_container.MythicCommandBase import *
import json


class XpcArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {
            "command": CommandParameter(
                name="command",
                type=ParameterType.ChooseOne,
                description="Choose an XPC command.",
                choices=[
                    "list",
                    "start",
                    "stop",
                    "load",
                    "unload",
                    "status",
                    "procinfo",
                    "submit",
                    "send",
                ],
                ui_position=1
            ),
            "program": CommandParameter(
                name="program",
                type=ParameterType.String,
                description="Program/binary to execute if using 'submit' command",
                required=False,
            ),
            "file": CommandParameter(
                name="Path to file to load on target",
                type=ParameterType.String,
                description="Path to the plist file if using load/unload commands",
                required=False,
            ),
            "servicename": CommandParameter(
                name="servicename",
                type=ParameterType.String,
                description="Name of the service to communicate with. Used with the submit, send, start/stop commands",
                required=False,
            ),
            "keepalive": CommandParameter(
                name="keepalive",
                type=ParameterType.Boolean,
                description="KeepAlive boolean",
                required=False,
            ),
            "pid": CommandParameter(
                name="pid",
                type=ParameterType.Number,
                description="PID of the process",
                required=False,
            ),
            "data": CommandParameter(
                name="data",
                type=ParameterType.String,
                description="base64 encoded json data to send to a target service",
                required=False,
            ),
        }

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)


class XpcCommand(CommandBase):
    cmd = "xpc"
    needs_admin = False
    help_cmd = "xpc"
    description = "Use xpc to execute routines with launchd or communicate with another service/process."
    version = 1
    author = "@xorrior"
    argument_class = XpcArguments
    attributes = CommandAttributes(
        # uncomment when poseidon can dynamically compile commands
        supported_os=[SupportedOS.MacOS]
    )
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
