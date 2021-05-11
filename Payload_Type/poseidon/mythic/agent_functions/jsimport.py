from mythic_payloadtype_container.MythicCommandBase import *
import base64
import sys
import json
from mythic_payloadtype_container.MythicRPC import *


class JsImportArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {
            "file_id": CommandParameter(
                name="JXA Script to Load",
                type=ParameterType.File,
                description="Select the JXA Script to load into memory",
                ui_position=1
            ),
        }

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)


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
        original_file_name = json.loads(task.original_params)["JXA Script to Load"]
        response = await MythicRPC().execute("create_file", task_id=task.id,
            file=base64.b64encode(task.args.get_arg("file_id")).decode(),
            saved_file_name=original_file_name,
            delete_after_fetch=True,
        )
        if response.status == MythicStatus.Success:
            task.args.add_arg("file_id", response.response["agent_file_id"])
            task.display_params = "script " + original_file_name
        else:
            raise Exception("Error from Mythic: " + response.error)
        return task

    async def process_response(self, response: AgentResponse):
        pass