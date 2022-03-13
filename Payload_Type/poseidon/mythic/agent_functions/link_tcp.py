from mythic_payloadtype_container.MythicCommandBase import *
import json


class LinkTCPArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="address",
                type=ParameterType.String,
                description="Address of the computer to connect to",
                parameter_group_info=[
                    ParameterGroupInfo(ui_position=1)
                ]
            ),
            CommandParameter(
                name="port",
                type=ParameterType.Number,
                description="Port to connect to",
                parameter_group_info=[
                    ParameterGroupInfo(ui_position=2)
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
                raise Exception("Failed to load arguments as JSON. Did you use the popup?")

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class LinkTCPCommand(CommandBase):
    cmd = "link_tcp"
    needs_admin = False
    help_cmd = "link_tcp {IP | host} {port}"
    description = "Link one agent to another over TCP."
    version = 1
    author = "@its_a_feature_"
    argument_class = LinkTCPArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
