from mythic_payloadtype_container.MythicCommandBase import *
import base64


class CurlArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="url",
                type=ParameterType.String,
                description="URL to request.",
                default_value="https://www.google.com",
                parameter_group_info=[
                    ParameterGroupInfo(ui_position=1)
                ]
            ),
            CommandParameter(
                name="method",
                type=ParameterType.ChooseOne,
                description="Type of request",
                choices=["GET", "POST"],
                parameter_group_info=[
                    ParameterGroupInfo(ui_position=2)
                ]
            ),
            CommandParameter(
                name="headers",
                type=ParameterType.String,
                description="JSON string of headers: {\"Host\":\"a.b.c.com\"}",
                parameter_group_info=[
                    ParameterGroupInfo(ui_position=3, required=False)
                ]
            ),
            CommandParameter(
                name="body",
                type=ParameterType.String,
                description="base64 encoded body.",
                parameter_group_info=[
                    ParameterGroupInfo(ui_position=4, required=False)
                ]
            ),
        ]

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise Exception("Must provide arguments")
        else:
            try:
                self.load_args_from_json_string(self.command_line)
            except:
                raise Exception("Failed to process arguments as JSON. Did you use the popup?")

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)
        if self.get_arg("headers") is not None:
            self.add_arg("headers", base64.b64encode(self.get_arg("headers").encode()).decode())


class CurlCommand(CommandBase):
    cmd = "curl"
    needs_admin = False
    help_cmd = 'curl {  "url": "https://www.google.com",  "method": "GET",  "headers": "",  "body": "" }'
    description = "Execute a single web request."
    version = 1
    author = "@xorrior"
    argument_class = CurlArguments
    attackmapping = ["T1213"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
