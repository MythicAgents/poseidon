from mythic_payloadtype_container.MythicCommandBase import *
import json


class CurlArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {
            "url": CommandParameter(
                name="url",
                type=ParameterType.String,
                description="URL to request.",
                default_value="https://www.google.com",
                ui_position=1
            ),
            "method": CommandParameter(
                name="method",
                type=ParameterType.ChooseOne,
                description="Type of request",
                choices=["GET", "POST"],
                ui_position=2
            ),
            "headers": CommandParameter(
                name="headers",
                type=ParameterType.String,
                description="base64 encoded json with headers.",
                required=False,
                ui_position=3
            ),
            "body": CommandParameter(
                name="Base64 body content",
                type=ParameterType.String,
                description="base64 encoded body.",
                required=False,
                ui_position=4
            ),
        }

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise Exception("Must provide arguments")
        else:
            try:
                self.load_args_from_json_string(self.command_line)
            except:
                raise Exception("Failed to process arguments as JSON. Did you use the popup?")


class CurlCommand(CommandBase):
    cmd = "curl"
    needs_admin = False
    help_cmd = 'curl {  "url": "https://www.google.com",  "method": "GET",  "headers": "",  "body": "" }'
    description = "Execute a single web request."
    version = 1
    author = "@xorrior"
    argument_class = CurlArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
