from mythic_payloadtype_container.MythicCommandBase import *
import base64
import sys
import json
from mythic_payloadtype_container.MythicRPC import *


class ExecuteMachoArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="file_id",
                display_name="Binary to execute",
                type=ParameterType.File,
                description="Select the Binary to execute in memory",
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
                description="Arguments to pass to binary",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=2
                    )
                ]
            ),
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class ExecuteMachoCommand(CommandBase):
    cmd = "execute_macho"
    needs_admin = False
    help_cmd = "execute_macho"
    description = "Upload a thin x64_Mach-o binary into memory and execute a function in-proc"
    version = 1
    author = ""
    argument_class = ExecuteMachoArguments
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
                                  comment="Uploaded into memory for execute_macho")
        task.display_params = f"{original_file_name} with args: {task.args.get_arg('args')}"
        return task

    async def process_response(self, response: AgentResponse):
        pass