"""
    Gitea API.

    This documentation describes the Gitea API.  # noqa: E501

    The version of the OpenAPI document: 1.19.0+dev-367-g8042ec1b9
    Generated by: https://openapi-generator.tech
"""


import sys
import unittest

import phab_tool.gitea_api
from phab_tool.gitea_api.model.commit_date_options import CommitDateOptions
from phab_tool.gitea_api.model.identity import Identity
globals()['CommitDateOptions'] = CommitDateOptions
globals()['Identity'] = Identity
from phab_tool.gitea_api.model.delete_file_options import DeleteFileOptions


class TestDeleteFileOptions(unittest.TestCase):
    """DeleteFileOptions unit test stubs"""

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def testDeleteFileOptions(self):
        """Test DeleteFileOptions"""
        # FIXME: construct object with mandatory attributes with example values
        # model = DeleteFileOptions()  # noqa: E501
        pass


if __name__ == '__main__':
    unittest.main()
