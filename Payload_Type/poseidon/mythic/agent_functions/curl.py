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
                description="Body contents to send in request",
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
                self.add_arg("method", self.get_arg("method").upper())
                if self.get_arg("headers") is not None:
                    self.add_arg("headers", base64.b64encode(self.get_arg("headers").encode()).decode())
                else:
                    self.add_arg("headers", "")
                if self.get_arg("body") is not None:
                    self.add_arg("body", base64.b64encode(self.get_arg("body").encode()).decode())
                else:
                    self.add_arg("body", "")
            except:
                raise Exception("Failed to process arguments as JSON.")

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)
        self.add_arg("method", self.get_arg("method").upper())
        if self.get_arg("headers") is not None:
            self.add_arg("headers", base64.b64encode(self.get_arg("headers").encode()).decode())
        else:
            self.add_arg("headers", "")
        if self.get_arg("body") is not None:
            self.add_arg("body", base64.b64encode(self.get_arg("body").encode()).decode())
        else:
            self.add_arg("body", "")


class CurlCommand(CommandBase):
    cmd = "curl"
    needs_admin = False
    help_cmd = 'curl -url https://www.google.com -method GET'
    description = "Execute a single web request."
    version = 1
    author = "@xorrior"
    argument_class = CurlArguments
    attackmapping = ["T1213"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        task.display_params = task.args.get_arg("url") + " via HTTP " + task.args.get_arg("method")
        if len(task.args.get_arg("headers")) > 0:
            task.display_params += " with custom headers"
        if len(task.args.get_arg("body")) > 0:
            task.display_params += " with custom body"
        return task

    async def process_response(self, response: AgentResponse):
        pass
