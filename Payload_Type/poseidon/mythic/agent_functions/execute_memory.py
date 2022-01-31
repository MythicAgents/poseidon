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
    attackmapping = ["T1106", "T1620", "T1105"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        file_resp = await MythicRPC().execute("get_file",
                                              file_id=task.args.get_arg("file_id"),
                                              task_id=task.id,
                                              get_contents=False)
        if file_resp.status == MythicRPCStatus.Success:
            original_file_name = file_resp.response[0]["filename"]
        else:
            raise Exception("Error from Mythic: " + str(file_resp.error))
        await MythicRPC().execute("update_file",
                                  file_id=task.args.get_arg("file_id"),
                                  delete_after_fetch=True,
                                  comment="Uploaded into memory for execute_memory")
        task.display_params = "function " + task.args.get_arg("function_name") + " of " + original_file_name + " with args: " + task.args.get_arg("args")
        return task

    async def process_response(self, response: AgentResponse):
        pass