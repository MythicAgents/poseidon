from mythic_payloadtype_container.MythicCommandBase import *
import base64
import sys
import json
from mythic_payloadtype_container.MythicRPC import *


class JsImportArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="file_id",
                display_name="JXA Script to Load",
                type=ParameterType.File,
                description="Select the JXA Script to load into memory",
            ),
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class JsImportCommand(CommandBase):
    cmd = "jsimport"
    needs_admin = False
    help_cmd = "jsimport"
    description = "Upload a script into memory for use with jsimport_call"
    version = 1
    author = "@its_a_feature_"
    argument_class = JsImportArguments
    attributes = CommandAttributes(
        # uncomment when poseidon can dynamically compile commands
        supported_os=[SupportedOS.MacOS]
    )
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        file_resp = await MythicRPC().execute("get_file",
                                              file_id=task.args.get_arg("file_id"),
                                              task_id=task.id,
                                              get_contents=False)
        if file_resp.status == MythicRPCStatus.Success:
            original_file_name = file_resp.response[0]["filename"]
        else:
            raise Exception("Error from Mythic: " + str(file_resp.error))
        task.display_params = f"script {original_file_name}"
        file_resp = await MythicRPC().execute("update_file",
                                              file_id=task.args.get_arg("file_id"),
                                              delete_after_fetch=True,
                                              comment="Uploaded into memory for jsimport")
        return task

    async def process_response(self, response: AgentResponse):
        pass