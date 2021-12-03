from mythic_payloadtype_container.MythicCommandBase import *
import json
import base64
import sys
from mythic_payloadtype_container.MythicRPC import *


class UploadArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="remote_path",
                display_name="Remote Path",
                type=ParameterType.String,
                description="Path where the uploaded file will be written.",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=2
                    )
                ]
            ),
            CommandParameter(
                name="file_id",
                display_name="File to Upload",
                type=ParameterType.File,
                description="The file to be written to the remote path.",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=1
                    )
                ]
            ),
            CommandParameter(
                name="overwrite",
                display_name="Overwrite Exiting File",
                type=ParameterType.Boolean,
                description="Overwrite file if it exists.",
                default_value=False,
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=3
                    )
                ]
             )
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class UploadCommand(CommandBase):
    cmd = "upload"
    needs_admin = False
    help_cmd = "upload"
    description = "upload a file to the target."
    version = 1
    supported_ui_features = ["file_browser:upload"]
    author = "@xorrior"
    argument_class = UploadArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        try:
            original_file_name = json.loads(task.original_params)["File to Upload"]
            if len(task.args.get_arg("remote_path")) == 0:
                task.args.add_arg("remote_path", original_file_name)
            elif task.args.get_arg("remote_path")[-1] == "/":
                task.args.add_arg("remote_path", task.args.get_arg("remote_path") + original_file_name)
            file_resp = await MythicRPC().execute("create_file", task_id=task.id,
                file=base64.b64encode(task.args.get_arg("file_id")).decode(),
                saved_file_name=original_file_name,
                delete_after_fetch=False,
            )
            if file_resp.status == MythicStatus.Success:
                task.args.add_arg("file_id", file_resp.response["agent_file_id"])
                task.display_params = f"{original_file_name} to {task.args.get_arg('remote_path')}"
            else:
                raise Exception("Error from Mythic: " + str(file_resp.error))
        except Exception as e:
            raise Exception("Error from Mythic: " + str(sys.exc_info()[-1].tb_lineno) + str(e))
        return task

    async def process_response(self, response: AgentResponse):
        pass
