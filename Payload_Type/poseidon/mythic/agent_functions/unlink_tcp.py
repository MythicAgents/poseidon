from mythic_payloadtype_container.MythicCommandBase import *
import json


class UnLinkTCPArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="connection",
                type=ParameterType.LinkInfo,
                description="Connection info for unlinking",
                parameter_group_info=[
                    ParameterGroupInfo(ui_position=1)
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


class UnLinkTCPCommand(CommandBase):
    cmd = "unlink_tcp"
    needs_admin = False
    help_cmd = "unlink_tcp"
    description = "Unlink a tcp connection."
    version = 1
    author = "@its_a_feature_"
    argument_class = UnLinkTCPArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        connection_info = task.args.get_arg("connection")
        remote_agent_id = connection_info.pop("callback_uuid", None)
        if remote_agent_id is None or remote_agent_id == "":
            raise Exception("Missing callback UUID in connection information")
        task.args.remove_arg("connection")
        task.args.add_arg("connection", remote_agent_id, type=ParameterType.String)
        task.display_params = f"from {remote_agent_id}"
        return task

    async def process_response(self, response: AgentResponse):
        pass
