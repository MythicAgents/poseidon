from mythic_payloadtype_container.MythicCommandBase import *
import base64
import sys
import json
from mythic_payloadtype_container.MythicRPC import *


class ExecuteMemoryArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="function_name",
                type=ParameterType.String,
                description="Which function should be executed?",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=2
                    )
                ]
            ),
            CommandParameter(
                name="file_id",
                display_name="Binary/Bundle to execute",
                type=ParameterType.File,
                description="Select the Bundle/Dylib/Binary to execute in memory",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=1
                    )
                ]
            ),
            CommandParameter(
                name="args",
                display_name="Argument String",
                type=ParameterType.String,
                description="Arguments to pass to function",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=3
                    )
                ]
            ),
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class ExecuteMemoryCommand(CommandBase):
    cmd = "execute_memory"
    needs_admin = False
    help_cmd = "execute_memory"
    description = "Upload a binary into memory and execute a function in-proc"
    version = 1
    author = "@its_a_feature_"
    argument_class = ExecuteMemoryArguments
    attributes = CommandAttributes(
        # uncomment when poseidon can dynamically compile commands
        supported_os=[SupportedOS.MacOS]
    )
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        original_file_name = json.loads(task.original_params)["Binary/Bundle to execute"]
        response = await MythicRPC().execute("create_file", task_id=task.id,
            file=base64.b64encode(task.args.get_arg("file_id")).decode(),
            saved_file_name=original_file_name,
            delete_after_fetch=True,
        )
        if response.status == MythicStatus.Success:
            task.args.add_arg("file_id", response.response["agent_file_id"])
            task.display_params = "function " + task.args.get_arg("function_name") + " of " + original_file_name + " with args: " + task.args.get_arg("args")
        else:
            raise Exception("Error from Mythic: " + response.error)
        return task

    async def process_response(self, response: AgentResponse):
        pass