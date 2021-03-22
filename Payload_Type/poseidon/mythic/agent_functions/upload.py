from mythic_payloadtype_container.MythicCommandBase import *
import json
from mythic_payloadtype_container.MythicFileRPC import *


class UploadArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {
            "remote_path": CommandParameter(
                name="Remote Path",
                type=ParameterType.String,
                description="Path where the uploaded file will be written.",
            ),
            "file_id": CommandParameter(
                name="File to Upload",
                type=ParameterType.File,
                description="The file to be written to the remote path.",
            ),
            "overwrite": CommandParameter(
                 name="Overwrite Exiting File",
                 type=ParameterType.Boolean,
                 description="Overwrite file if it exists.",
                 default_value=False
             )
        }

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)


class UploadCommand(CommandBase):
    cmd = "upload"
    needs_admin = False
    help_cmd = "upload"
    description = "upload a file to the target."
    version = 1
    is_exit = False
    is_file_browse = False
    is_process_list = False
    is_download_file = False
    is_remove_file = False
    is_upload_file = True
    author = "@xorrior"
    argument_class = UploadArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        original_file_name = json.loads(task.original_params)["File to Upload"]
        response = await MythicFileRPC(task).register_file(
            file=task.args.get_arg("file_id"),
            saved_file_name=original_file_name,
            delete_after_fetch=False,
        )
        if response.status == MythicStatus.Success:
            task.args.add_arg("file_id", response.agent_file_id)
        else:
            raise Exception("Error from Mythic: " + response.error_message)
        return task

    async def process_response(self, response: AgentResponse):
        pass
